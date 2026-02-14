package grpc

import (
	"context"
	"errors"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"tasks/logger"
	"tasks/proto/taskpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func ApplyTaskFieldsFromRequestLocal(task *persistence.Task, req interface{}) error {
	switch r := req.(type) {
	case *taskpb.CreateTaskRequest:
		task.Title = r.Title
		task.Description = r.Description
		task.PerformerId = uint(r.PerformerId)
		task.CreatorId = uint(r.CreatorId)
		task.Observers = persistence.ObserversFromIDs(r.ObserverIds)
		task.Status = r.Status
		// CreatedAt/UpdatedAt should be set by persistence/usecase as appropriate
	case *taskpb.UpdateTaskRequest:
		task.Title = r.Title
		task.Description = r.Description
		task.PerformerId = uint(r.PerformerId)
		task.CreatorId = uint(r.CreatorId)
		task.Observers = persistence.ObserversFromIDs(r.ObserverIds)
		task.Status = r.Status
	default:
		return gorm.ErrInvalidField
	}
	return nil
}

func convertToProtoLocal(task persistence.Task, shard *gorm.DB) *taskpb.Task {
	return &taskpb.Task{
		Id:          uint64(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		PerformerId: uint64(task.PerformerId),
		CreatorId:   uint64(task.CreatorId),
		ObserverIds: task.ObserverIDs(shard),
	}
}

func (s *TaskServer) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.TaskResponse, error) {
	taskID := uint(req.Id)
	currentShardIndex, err := cache.GetTaskShard(ctx, taskID)
	var fromShard *gorm.DB
	var task persistence.Task

	if err == nil {
		fromShard = s.ShardManager.GetShardByIndex(currentShardIndex)
		if fromShard != nil {
			if err := fromShard.First(&task, req.Id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
				}
				return nil, err
			}
		}
	} else {
		if cache.IsNilError(err) {
			logger.Warn(ctx, "cache shard mapping miss for update", logger.ZapUint("task_id", taskID))
		} else {
			logger.Warn(ctx, "cache error on get shard for update", logger.ZapError(err))
		}
	}

	// If cache didn't provide shard or DB lookup failed, fallback to scanning shards
	if fromShard == nil {
		allShards := s.ShardManager.GetAllShards()
		found := false
		for idx, sh := range allShards {
			if sh == nil {
				continue
			}
			var t persistence.Task
			// use silent GORM logger for probe query to avoid noisy "record not found" logs
			if err := sh.Session(&gorm.Session{Logger: glogger.Default.LogMode(glogger.Silent)}).First(&t, req.Id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			// found
			fromShard = sh
			currentShardIndex = idx
			task = t
			found = true
			break
		}
		if !found {
			return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
		}
	}

	oldPerformerID := task.PerformerId

	if err := fromShard.Where("task_id = ?", task.ID).Delete(&persistence.Observer{}).Error; err != nil {
		return nil, err
	}

	if err := ApplyTaskFieldsFromRequestLocal(&task, req); err != nil {
		return nil, err
	}

	newShardIndex := s.ShardManager.GetShardByPerformerIDIndex(task.PerformerId)
	needMigrate := oldPerformerID != task.PerformerId && newShardIndex != currentShardIndex

	var shard *gorm.DB
	if needMigrate {
		toShard := s.ShardManager.GetShardByIndex(newShardIndex)
		if toShard == nil {
			return nil, status.Errorf(codes.NotFound, "target shard not found for performer %d", task.PerformerId)
		}
		if err := migrateTaskToShard(ctx, &task, fromShard, toShard, currentShardIndex, newShardIndex); err != nil {
			return nil, err
		}
		shard = toShard
	} else {
		if err := fromShard.Save(&task).Error; err != nil {
			return nil, err
		}
		for _, obs := range task.Observers {
			newObs := persistence.Observer{UserId: obs.UserId, TaskId: task.ID}
			if err := fromShard.Create(&newObs).Error; err != nil {
				return nil, err
			}
		}
		shard = fromShard
	}

	if err := cache.DeleteTaskCache(ctx, task.ID); err != nil {
		logger.Warn(ctx, "cache delete failed", logger.ZapError(err))
	}

	if s.Producer != nil {
		if err := s.Producer.PublishTaskEvent(ctx, "TaskUpdated", task, shard); err != nil {
			return nil, err
		}
	}

	return &taskpb.TaskResponse{Task: convertToProtoLocal(task, shard)}, nil
}
