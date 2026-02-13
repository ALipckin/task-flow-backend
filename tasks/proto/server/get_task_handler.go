package server

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

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	taskID := uint(req.Id)
	task, err := cache.CacheGetTask(ctx, taskID)
	if err == nil {
		shardIndex, _ := cache.GetTaskShard(ctx, taskID)
		shard := s.ShardManager.GetShardByIndex(shardIndex)
		if shard != nil {
			return &taskpb.TaskResponse{Task: convertToProto(*task, shard)}, nil
		}
	} else {
		// if cache miss, log a warning but continue with DB lookup
		if cache.IsNilError(err) {
			logger.Warn(ctx, "cache miss for task", logger.ZapUint("task_id", taskID))
		} else {
			logger.Warn(ctx, "cache error on get", logger.ZapError(err))
		}
	}

	// Try to get shard index from cache; if not present, fall back to scanning shards.
	shardIndex, err := cache.GetTaskShard(ctx, taskID)
	if err == nil {
		shard := s.ShardManager.GetShardByIndex(shardIndex)
		if shard == nil {
			return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
		}
		task = &persistence.Task{}
		if err := shard.First(task, req.Id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
			}
			return nil, err
		}

		if err := cache.CacheSetTask(ctx, *task); err != nil {
			// don't fail the request purely because of cache write error
			logger.Warn(ctx, "cache set failed", logger.ZapError(err))
		}

		return &taskpb.TaskResponse{Task: convertToProto(*task, shard)}, nil
	}

	// Fallback: scan all shards to find the task (in case cache shard mapping was missing)
	allShards := s.ShardManager.GetAllShards()
	for _, shard := range allShards {
		var t persistence.Task
		// use silent GORM logger for probe query to avoid noisy "record not found" logs
		if err := shard.Session(&gorm.Session{Logger: glogger.Default.LogMode(glogger.Silent)}).First(&t, req.Id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			// unexpected DB error
			return nil, err
		}
		// Found task in this shard; try to set cache but don't fail the request on cache errors
		if err := cache.CacheSetTask(ctx, t); err != nil {
			logger.Warn(ctx, "cache set failed", logger.ZapError(err))
		}
		return &taskpb.TaskResponse{Task: convertToProto(t, shard)}, nil
	}

	return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
}
