package adapters

import (
	"context"
	"tasks/internal/infrastructure/cache"
)

type RedisIDAllocator struct{}

func NewRedisIDAllocator() *RedisIDAllocator { return &RedisIDAllocator{} }

func (a *RedisIDAllocator) NextID(ctx context.Context) (uint, error) {
	return cache.AllocTaskID(ctx)
}
