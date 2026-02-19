package use_case

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/ports"
	"tasks/logger"
)

type GetTask struct {
	repo     ports.Repository
	cache    ports.Cache
	producer ports.EventProducer
}

// NewGetTask constructs GetTask use-case with its dependencies.
func NewGetTask(
	repo ports.Repository,
	cache ports.Cache,
	producer ports.EventProducer,
) *GetTask {
	return &GetTask{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

type GetTaskCommand struct {
	ID uint64
}

// Execute returns tasks matching the command filter across all shards.
func (uc *GetTask) Execute(ctx context.Context, cmd GetTaskCommand) (domain.Task, error) {
	taskID := uint(cmd.ID)

	// Try cache
	task, err := uc.cache.GetTask(ctx, taskID)
	if err == nil {
		return task, nil
	}

	logger.Warn(ctx, "Cache not found for task",
		logger.ZapUint("task_id", taskID),
		logger.ZapError(err),
	)

	// Fetch from repository
	repoTask, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	// Best-effort cache set
	_ = uc.cache.SetTask(ctx, *repoTask)

	return *repoTask, nil
}
