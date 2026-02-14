package grpc

import (
	"context"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/adapters"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"tasks/internal/use_case"
	"tasks/proto/taskpb"

	"gorm.io/gorm"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	ShardManager *shard.ShardManager
	CreateUC     *use_case.CreateTask
	GetTasksUC   *use_case.GetTasks
	Producer     *adapters.KafkaProducerAdapter
}

// migrateTaskToShard copies task and its observers to another shard, deletes from the old shard, updates Redis mapping.
func migrateTaskToShard(ctx context.Context, task *persistence.Task, fromShard, toShard *gorm.DB, fromIndex, toIndex int) error {
	if err := toShard.Create(task).Error; err != nil {
		return err
	}
	for _, obs := range task.Observers {
		newObs := persistence.Observer{UserId: obs.UserId, TaskId: task.ID}
		if err := toShard.Create(&newObs).Error; err != nil {
			return err
		}
	}
	if err := fromShard.Unscoped().Where("task_id = ?", task.ID).Delete(&persistence.Observer{}).Error; err != nil {
		return err
	}
	if err := fromShard.Unscoped().Delete(task).Error; err != nil {
		return err
	}
	if err := cache.SetTaskShard(ctx, task.ID, toIndex); err != nil {
		return err
	}
	// remove cached key via cache helper
	_ = cache.DeleteTaskCache(ctx, task.ID)
	return nil
}
