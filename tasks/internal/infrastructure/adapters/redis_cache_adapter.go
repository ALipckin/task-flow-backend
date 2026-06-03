package adapters

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
)

type RedisCacheAdapter struct {
	store *cache.Store
}

func NewRedisCacheAdapter(store *cache.Store) *RedisCacheAdapter {
	return &RedisCacheAdapter{store: store}
}

func (a *RedisCacheAdapter) SetTask(ctx context.Context, task domain.Task) error {
	return a.store.SetTask(ctx, persistence.TaskFromDomain(task))
}

func (a *RedisCacheAdapter) GetTask(ctx context.Context, taskID uint) (domain.Task, error) {
	p, err := a.store.GetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	return persistence.TaskToDomain(*p), nil
}

func (a *RedisCacheAdapter) DeleteTask(ctx context.Context, taskID uint) error {
	return a.store.DeleteTaskCache(ctx, taskID)
}
