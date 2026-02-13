package adapters

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
)

type RedisCacheAdapter struct{}

func NewRedisCacheAdapter() *RedisCacheAdapter { return &RedisCacheAdapter{} }

func (a *RedisCacheAdapter) Set(ctx context.Context, task domain.Task) error {
	p := persistence.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		PerformerId: task.PerformerId,
		CreatorId:   task.CreatorId,
		Status:      task.Status,
	}
	return cache.CacheSetTask(ctx, p)
}

func (a *RedisCacheAdapter) Get(ctx context.Context, taskID uint) (domain.Task, error) {
	p, err := cache.CacheGetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	return domain.Task{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		PerformerId: p.PerformerId,
		CreatorId:   p.CreatorId,
		Status:      p.Status,
	}, nil
}
