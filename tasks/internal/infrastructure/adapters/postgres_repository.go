package adapters

import (
	"context"
	"errors"
	"tasks/internal/domain"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"tasks/internal/ports"
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
		return err
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
		return nil, err
	}

	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return nil, errors.New("shard not found")
	}

	var task persistence.Task

	if err := db.WithContext(ctx).
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
