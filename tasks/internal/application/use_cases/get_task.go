package use_cases

import (
	"context"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

type GetTask struct {
	repo  out.Repository
	cache out.Cache
}

func NewGetTask(
	repo out.Repository,
	cache out.Cache,
) *GetTask {
	return &GetTask{
		repo:  repo,
		cache: cache,
	}
}

func (uc *GetTask) Execute(ctx context.Context, query queries.GetTaskQuery) (domain.Task, error) {
	taskID := uint(query.ID)

	task, err := uc.cache.GetTask(ctx, taskID)
	if err == nil {
		return task, nil
	}

	repoTask, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	_ = uc.cache.SetTask(ctx, *repoTask)

	return *repoTask, nil
}
