package grpc

import (
	"context"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"tasks/proto/taskpb"

	"gorm.io/gorm"
)

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	taskID := uint(req.Id)
	shardIndex, err := cache.GetTaskShard(ctx, taskID)
	if err != nil {
		return nil, err
	}
	shard := s.ShardManager.GetShardByIndex(shardIndex)
	if shard == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var task persistence.Task
	if err := shard.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := shard.Delete(&task).Error; err != nil {
		return nil, err
	}

	_ = cache.DeleteTaskCache(ctx, task.ID)
	_ = cache.DelTaskShard(ctx, task.ID)

	if s.Producer != nil {
		if err := s.Producer.PublishTaskEvent(ctx, "TaskDeleted", task, shard); err != nil {
			return nil, err
		}
	}

	return &taskpb.DeleteTaskResponse{Message: "Task deleted"}, nil
}
