package adapters

import (
	"context"
	"tasks/internal/application/ports/out"
	"tasks/internal/infrastructure/cache"
)

type RedisTaskShardIndexAdapter struct{}

func NewRedisTaskShardIndexAdapter() *RedisTaskShardIndexAdapter {
	return &RedisTaskShardIndexAdapter{}
}

func (a *RedisTaskShardIndexAdapter) Get(ctx context.Context, taskID uint) (int, error) {
	idx, err := cache.GetTaskShard(ctx, taskID)
	if err != nil {
		if cache.IsNilError(err) {
			return -1, out.ErrNotFound
		}
		return -1, err
	}
	return idx, nil
}

func (a *RedisTaskShardIndexAdapter) Set(ctx context.Context, taskID uint, shardIndex int) error {
	return cache.SetTaskShard(ctx, taskID, shardIndex)
}

func (a *RedisTaskShardIndexAdapter) Delete(ctx context.Context, taskID uint) error {
	return cache.DelTaskShard(ctx, taskID)
}
