package adapters

import (
	"context"
	"tasks/internal/infrastructure/cache"
)

type RedisIDAllocator struct {
	store *cache.Store
}

func NewRedisIDAllocator(store *cache.Store) *RedisIDAllocator {
	return &RedisIDAllocator{store: store}
}

func (a *RedisIDAllocator) NextID(ctx context.Context) (uint, error) {
	return a.store.AllocTaskID(ctx)
}
