package adapters

import (
	"context"
	"tasks/internal/application/ports/out"
	"tasks/internal/infrastructure/cache"
)

type RedisTaskShardIndexAdapter struct {
	store *cache.Store
}

func NewRedisTaskShardIndexAdapter(store *cache.Store) *RedisTaskShardIndexAdapter {
	return &RedisTaskShardIndexAdapter{store: store}
}

func (a *RedisTaskShardIndexAdapter) Get(ctx context.Context, taskID uint) (int, error) {
	idx, err := a.store.GetTaskShard(ctx, taskID)
	if err != nil {
		if cache.IsNilError(err) {
			return -1, out.ErrNotFound
		}
		return -1, err
	}
	return idx, nil
}

func (a *RedisTaskShardIndexAdapter) Set(ctx context.Context, taskID uint, shardIndex int) error {
	return a.store.SetTaskShard(ctx, taskID, shardIndex)
}

func (a *RedisTaskShardIndexAdapter) Delete(ctx context.Context, taskID uint) error {
	return a.store.DelTaskShard(ctx, taskID)
}
