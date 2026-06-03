package use_cases

import (
	"context"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
	"tasks/logger"
)

type GetTask struct {
	repo     out.Repository
	cache    out.Cache
	producer out.EventProducer
	log      *logger.Logger
}

func NewGetTask(
	repo out.Repository,
	cache out.Cache,
	producer out.EventProducer,
	log *logger.Logger,
) *GetTask {
	return &GetTask{
		repo:     repo,
		cache:    cache,
		producer: producer,
		log:      log,
	}
}

func (uc *GetTask) Execute(ctx context.Context, query queries.GetTaskQuery) (domain.Task, error) {
	taskID := uint(query.ID)

	task, err := uc.cache.GetTask(ctx, taskID)
	if err == nil {
		return task, nil
	}

	uc.log.Warn(ctx, "Cache not found for task",
		logger.ZapUint("task_id", taskID),
		logger.ZapError(err),
	)

	repoTask, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	_ = uc.cache.SetTask(ctx, *repoTask)

	return *repoTask, nil
}
