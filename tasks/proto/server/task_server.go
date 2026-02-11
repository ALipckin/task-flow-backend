package server

import (
	"context"
	"tasks/helpers"
	"tasks/initializers"
	"tasks/models"
	"tasks/proto/taskpb"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	ShardManager *models.ShardManager
}

func (s *TaskServer) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskResponse, error) {
	var task models.Task
	if err := helpers.ApplyTaskFieldsFromRequest(&task, req); err != nil {
		return nil, err
	}

	id, err := helpers.AllocTaskID(ctx)
	if err != nil {
		return nil, err
	}
	task.ID = id

	// Shard by performer_id (consistent hash ring)
	shard := s.ShardManager.GetShardByPerformerID(task.PerformerId)
	shardIndex := s.ShardManager.GetShardByPerformerIDIndex(task.PerformerId)
	if err := shard.Create(&task).Error; err != nil {
		return nil, err
	}

	if err := helpers.SetTaskShard(ctx, task.ID, shardIndex); err != nil {
		return nil, err
	}
	if err := helpers.CacheSetTask(ctx, task); err != nil {
		return nil, err
	}

	if err := helpers.SendTaskEventToKafka("TaskCreated", task, shard); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task, shard)}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	taskID := uint(req.Id)
	task, err := helpers.CacheGetTask(ctx, taskID)
	if err == nil {
		shardIndex, _ := helpers.GetTaskShard(ctx, taskID)
		shard := s.ShardManager.GetShardByIndex(shardIndex)
		if shard != nil {
			return &taskpb.TaskResponse{Task: convertToProto(*task, shard)}, nil
		}
	}

	shardIndex, err := helpers.GetTaskShard(ctx, taskID)
	if err != nil {
		return nil, err
	}
	shard := s.ShardManager.GetShardByIndex(shardIndex)
	if shard == nil {
		return nil, gorm.ErrRecordNotFound
	}
	task = &models.Task{}
	if err := shard.First(task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := helpers.CacheSetTask(ctx, *task); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(*task, shard)}, nil
}

func (s *TaskServer) GetTasks(ctx context.Context, req *taskpb.GetTasksRequest) (*taskpb.GetTasksResponse, error) {
	allShards := s.ShardManager.GetAllShards()
	protoTasks := make([]*taskpb.Task, 0)

	for _, shard := range allShards {
		var tasks []models.Task
		query := shard

		if req.Title != "" {
			query = query.Where("title = ?", req.Title)
		}
		if req.CreatorId != 0 {
			query = query.Where("creator_id = ?", req.CreatorId)
		}
		if req.PerformerId != 0 {
			query = query.Where("performer_id = ?", req.PerformerId)
		}

		if err := query.Find(&tasks).Error; err != nil {
			continue
		}
		for _, t := range tasks {
			protoTasks = append(protoTasks, convertToProto(t, shard))
		}
	}

	return &taskpb.GetTasksResponse{Tasks: protoTasks}, nil
}

func (s *TaskServer) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.TaskResponse, error) {
	taskID := uint(req.Id)
	currentShardIndex, err := helpers.GetTaskShard(ctx, taskID)
	if err != nil {
		return nil, err
	}
	fromShard := s.ShardManager.GetShardByIndex(currentShardIndex)
	if fromShard == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var task models.Task
	if err := fromShard.First(&task, req.Id).Error; err != nil {
		return nil, err
	}
	oldPerformerID := task.PerformerId

	if err := fromShard.Where("task_id = ?", task.ID).Delete(&models.Observer{}).Error; err != nil {
		return nil, err
	}

	if err := helpers.ApplyTaskFieldsFromRequest(&task, req); err != nil {
		return nil, err
	}

	newShardIndex := s.ShardManager.GetShardByPerformerIDIndex(task.PerformerId)
	needMigrate := oldPerformerID != task.PerformerId && newShardIndex != currentShardIndex

	var shard *gorm.DB
	if needMigrate {
		toShard := s.ShardManager.GetShardByIndex(newShardIndex)
		if toShard == nil {
			return nil, gorm.ErrRecordNotFound
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
			newObs := models.Observer{UserId: obs.UserId, TaskId: task.ID}
			if err := fromShard.Create(&newObs).Error; err != nil {
				return nil, err
			}
		}
		shard = fromShard
	}

	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))

	if err := helpers.SendTaskEventToKafka("TaskUpdated", task, shard); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task, shard)}, nil
}

// migrateTaskToShard copies task and its observers to another shard, deletes from the old shard, updates Redis mapping.
func migrateTaskToShard(ctx context.Context, task *models.Task, fromShard, toShard *gorm.DB, fromIndex, toIndex int) error {
	if err := toShard.Create(task).Error; err != nil {
		return err
	}
	for _, obs := range task.Observers {
		newObs := models.Observer{UserId: obs.UserId, TaskId: task.ID}
		if err := toShard.Create(&newObs).Error; err != nil {
			return err
		}
	}
	if err := fromShard.Unscoped().Where("task_id = ?", task.ID).Delete(&models.Observer{}).Error; err != nil {
		return err
	}
	if err := fromShard.Unscoped().Delete(task).Error; err != nil {
		return err
	}
	if err := helpers.SetTaskShard(ctx, task.ID, toIndex); err != nil {
		return err
	}
	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))
	return nil
}

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	taskID := uint(req.Id)
	shardIndex, err := helpers.GetTaskShard(ctx, taskID)
	if err != nil {
		return nil, err
	}
	shard := s.ShardManager.GetShardByIndex(shardIndex)
	if shard == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var task models.Task
	if err := shard.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := shard.Delete(&task).Error; err != nil {
		return nil, err
	}

	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))
	_ = helpers.DelTaskShard(ctx, task.ID)

	if err := helpers.SendTaskEventToKafka("TaskDeleted", task, shard); err != nil {
		return nil, err
	}

	return &taskpb.DeleteTaskResponse{Message: "Task deleted"}, nil
}

func convertToProto(task models.Task, shard *gorm.DB) *taskpb.Task {
	return &taskpb.Task{
		Id:          uint64(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		PerformerId: uint64(task.PerformerId),
		CreatorId:   uint64(task.CreatorId),
		ObserverIds: task.ObserverIDs(shard),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}
