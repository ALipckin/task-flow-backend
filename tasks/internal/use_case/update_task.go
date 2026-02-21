package use_case

import (
	"context"
	"errors"
	"tasks/internal/domain"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/adapters"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"tasks/logger"

	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

type UpdateTask struct {
	sharder  *shard.ShardManager
	producer *adapters.KafkaProducerAdapter
}

func NewUpdateTask(
	sharder *shard.ShardManager,
	producer *adapters.KafkaProducerAdapter,
) *UpdateTask {
	return &UpdateTask{
		sharder:  sharder,
		producer: producer,
	}
}

type UpdateTaskCommand struct {
	ID          uint64
	Title       string
	Description string
	Status      string
	PerformerID uint
	CreatorID   uint
	ObserverIDs []uint64
}

func (uc *UpdateTask) Execute(ctx context.Context, cmd UpdateTaskCommand) (domain.Task, error) {
	taskID := uint(cmd.ID)

	currentShardIndex, err := cache.GetTaskShard(ctx, taskID)
	var fromShard *gorm.DB
	var task persistence.Task

	if err == nil {
		fromShard = uc.sharder.GetShardByIndex(currentShardIndex)
		if fromShard != nil {
			if err := fromShard.First(&task, cmd.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.Task{}, gorm.ErrRecordNotFound
				}
				return domain.Task{}, err
			}
		}
	} else {
		if cache.IsNilError(err) {
			logger.Warn(ctx, "cache shard mapping miss for update", logger.ZapUint("task_id", taskID))
		} else {
			logger.Warn(ctx, "cache error on get shard for update", logger.ZapError(err))
		}
	}

	if fromShard == nil {
		allShards := uc.sharder.GetAllShards()
		found := false

		for idx, sh := range allShards {
			if sh == nil {
				continue
			}

			var t persistence.Task
			if err := sh.Session(&gorm.Session{Logger: glogger.Default.LogMode(glogger.Silent)}).First(&t, cmd.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return domain.Task{}, err
			}

			fromShard = sh
			currentShardIndex = idx
			task = t
			found = true
			break
		}

		if !found {
			return domain.Task{}, gorm.ErrRecordNotFound
		}
	}

	oldPerformerID := task.PerformerId

	if err := fromShard.Where("task_id = ?", task.ID).Delete(&persistence.Observer{}).Error; err != nil {
		return domain.Task{}, err
	}

	task.Title = cmd.Title
	task.Description = cmd.Description
	task.PerformerId = cmd.PerformerID
	task.CreatorId = cmd.CreatorID
	task.Observers = persistence.ObserversFromIDs(cmd.ObserverIDs)
	task.Status = cmd.Status

	newShardIndex := uc.sharder.GetShardByPerformerIDIndex(task.PerformerId)
	needMigrate := oldPerformerID != task.PerformerId && newShardIndex != currentShardIndex

	var shardDB *gorm.DB
	if needMigrate {
		toShard := uc.sharder.GetShardByIndex(newShardIndex)
		if toShard == nil {
			return domain.Task{}, errors.New("target shard not found")
		}

		if err := migrateTaskToShard(ctx, &task, fromShard, toShard, newShardIndex); err != nil {
			return domain.Task{}, err
		}
		shardDB = toShard
	} else {
		if err := fromShard.Save(&task).Error; err != nil {
			return domain.Task{}, err
		}

		for _, obs := range task.Observers {
			newObs := persistence.Observer{UserId: obs.UserId, TaskId: task.ID}
			if err := fromShard.Create(&newObs).Error; err != nil {
				return domain.Task{}, err
			}
		}
		shardDB = fromShard
	}

	if err := cache.DeleteTaskCache(ctx, task.ID); err != nil {
		logger.Warn(ctx, "cache delete failed", logger.ZapError(err))
	}

	if uc.producer != nil {
		if err := uc.producer.PublishTaskEvent(ctx, "TaskUpdated", task, shardDB); err != nil {
			return domain.Task{}, err
		}
	}

	return domain.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		PerformerId: task.PerformerId,
		CreatorId:   task.CreatorId,
		Observers:   task.Observers,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func migrateTaskToShard(
	ctx context.Context,
	task *persistence.Task,
	fromShard *gorm.DB,
	toShard *gorm.DB,
	toIndex int,
) error {
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
	_ = cache.DeleteTaskCache(ctx, task.ID)

	return nil
}
