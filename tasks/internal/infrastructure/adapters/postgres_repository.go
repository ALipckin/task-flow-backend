package adapters

import (
	"context"
	"errors"
	"tasks/internal/domain"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"tasks/internal/ports"
	"tasks/logger"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// PostgresRepository implements ports.Repository using GORM shards via shard.Manager.
type PostgresRepository struct {
	ShardManager *shard.ShardManager
}

func NewPostgresRepository(sm *shard.ShardManager) *PostgresRepository {
	return &PostgresRepository{ShardManager: sm}
}

func (r *PostgresRepository) Save(ctx context.Context, t domain.Task, shardIndex int) error {
	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return errors.New("shard not found")
	}
	p := persistence.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		PerformerId: t.PerformerId,
		CreatorId:   t.CreatorId,
		Status:      t.Status,
	}
	return db.Create(&p).Error
}

// Find queries tasks in the specified shard index using the provided filter.
// If shardIndex is negative, returns an error (caller should iterate shards itself).
func (r *PostgresRepository) Find(ctx context.Context, filter ports.TaskFilter, shardIndex int) ([]domain.Task, error) {
	if shardIndex < 0 {
		return nil, errors.New("shard index required")
	}
	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return nil, errors.New("shard not found")
	}

	var models []persistence.Task
	query := db
	if filter.Title != "" {
		query = query.Where("title = ?", filter.Title)
	}
	if filter.CreatorID != 0 {
		query = query.Where("creator_id = ?", filter.CreatorID)
	}
	if filter.PerformerID != 0 {
		query = query.Where("performer_id = ?", filter.PerformerID)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]domain.Task, 0, len(models))
	for _, m := range models {
		result = append(result, domain.Task{
			ID:          m.ID,
			Title:       m.Title,
			Description: m.Description,
			PerformerId: m.PerformerId,
			CreatorId:   m.CreatorId,
			Observers:   m.Observers,
			Status:      m.Status,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			DeletedAt:   m.DeletedAt,
		})
	}

	return result, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, taskID uint) error {
	shardIndex, err := cache.GetTaskShard(ctx, taskID)
	if err != nil {
		if !cache.IsNilError(err) {
			return err
		}

		shardIndex, err = r.findShardIndexByTaskID(ctx, taskID)
		if err != nil {
			return err
		}
		_ = cache.SetTaskShard(ctx, taskID, shardIndex)
	}
	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return errors.New("shard not found")
	}

	return db.Delete(&persistence.Task{ID: taskID}, taskID).Error
}

func (r *PostgresRepository) GetByID(ctx context.Context, taskID uint) (*domain.Task, error) {
	shardIndex, err := cache.GetTaskShard(ctx, taskID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			for idx, db := range r.ShardManager.GetAllShards() {
				var task persistence.Task
				err := db.WithContext(ctx).
					Preload("Observers").
					First(&task, taskID).Error
				if err == nil {
					_ = cache.SetTaskShard(ctx, task.ID, idx)
					return &domain.Task{
						ID:          task.ID,
						Title:       task.Title,
						Description: task.Description,
						PerformerId: task.PerformerId,
						CreatorId:   task.CreatorId,
						Status:      task.Status,
						Observers:   task.Observers,
						CreatedAt:   task.CreatedAt,
						UpdatedAt:   task.UpdatedAt,
					}, nil
				}
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			return nil, gorm.ErrRecordNotFound // not found in any shard

		} else {
			return nil, err // real infrastructure failure
		}
	}
	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return nil, errors.New("shard not found")
	}

	var task persistence.Task

	if err := db.WithContext(ctx).
		Preload("Observers").
		First(&task, taskID).Error; err != nil {
		return nil, err
	}

	return &domain.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		PerformerId: task.PerformerId,
		CreatorId:   task.CreatorId,
		Status:      task.Status,
		Observers:   task.Observers,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (r *PostgresRepository) findShardIndexByTaskID(ctx context.Context, taskID uint) (int, error) {
	for idx, db := range r.ShardManager.GetAllShards() {
		if db == nil {
			continue
		}

		var task persistence.Task
		err := db.WithContext(ctx).
			Session(&gorm.Session{Logger: glogger.Default.LogMode(glogger.Silent)}).
			Select("id").
			First(&task, taskID).Error
		if err == nil {
			return idx, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}

		return -1, err
	}

	return -1, gorm.ErrRecordNotFound
}

func (r *PostgresRepository) Update(ctx context.Context, input ports.UpdateTaskInput) (*domain.Task, error) {
	taskID := input.ID

	currentShardIndex, err := cache.GetTaskShard(ctx, taskID)
	var fromShard *gorm.DB
	var task persistence.Task

	if err == nil {
		fromShard = r.ShardManager.GetShardByIndex(currentShardIndex)
		if fromShard != nil {
			if err := fromShard.First(&task, taskID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, gorm.ErrRecordNotFound
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

	if fromShard == nil {
		allShards := r.ShardManager.GetAllShards()
		found := false
		for idx, sh := range allShards {
			if sh == nil {
				continue
			}

			var t persistence.Task
			if err := sh.Session(&gorm.Session{Logger: glogger.Default.LogMode(glogger.Silent)}).First(&t, taskID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}

			fromShard = sh
			currentShardIndex = idx
			task = t
			found = true
			break
		}
		if !found {
			return nil, gorm.ErrRecordNotFound
		}
	}

	oldPerformerID := task.PerformerId

	if err := fromShard.Where("task_id = ?", task.ID).Delete(&persistence.Observer{}).Error; err != nil {
		return nil, err
	}

	task.Title = input.Title
	task.Description = input.Description
	task.PerformerId = input.PerformerID
	task.CreatorId = input.CreatorID
	task.Observers = observersFromUintIDs(input.ObserverIDs)
	task.Status = input.Status

	newShardIndex := r.ShardManager.GetShardByPerformerIDIndex(task.PerformerId)
	needMigrate := oldPerformerID != task.PerformerId && newShardIndex != currentShardIndex

	if needMigrate {
		toShard := r.ShardManager.GetShardByIndex(newShardIndex)
		if toShard == nil {
			return nil, errors.New("target shard not found")
		}

		if err := migrateTaskToShard(ctx, &task, fromShard, toShard, newShardIndex); err != nil {
			return nil, err
		}
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
	}

	if err := cache.DeleteTaskCache(ctx, task.ID); err != nil {
		logger.Warn(ctx, "cache delete failed", logger.ZapError(err))
	}

	return persistenceToDomainTask(task), nil
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

func observersFromUintIDs(ids []uint) []persistence.Observer {
	if len(ids) == 0 {
		return nil
	}

	observers := make([]persistence.Observer, len(ids))
	for i := range ids {
		observers[i] = persistence.Observer{UserId: ids[i]}
	}
	return observers
}

func persistenceToDomainTask(task persistence.Task) *domain.Task {
	return &domain.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		PerformerId: task.PerformerId,
		CreatorId:   task.CreatorId,
		Status:      task.Status,
		Observers:   task.Observers,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		DeletedAt:   task.DeletedAt,
	}
}
