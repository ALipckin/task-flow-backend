package adapters

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
)

type RedisCacheAdapter struct{}

func NewRedisCacheAdapter() *RedisCacheAdapter { return &RedisCacheAdapter{} }

func (a *RedisCacheAdapter) SetTask(ctx context.Context, task domain.Task) error {
	return cache.SetTask(ctx, persistence.TaskFromDomain(task))
}

func (a *RedisCacheAdapter) GetTask(ctx context.Context, taskID uint) (domain.Task, error) {
	p, err := cache.GetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	return persistence.TaskToDomain(*p), nil
}

func (a *RedisCacheAdapter) DeleteTask(ctx context.Context, taskID uint) error {
	return cache.DeleteTaskCache(ctx, taskID)
}
