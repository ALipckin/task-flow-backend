package adapters

import (
	"context"
	"errors"
	"tasks/internal/domain"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/persistence"
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
