package queries

import (
	"context"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
	"tasks/logger"
)

type GetTask struct {
	repo     out.Repository
	cache    out.Cache
	producer out.EventProducer
}

// NewGetTask constructs GetTask use-case with its dependencies.
func NewGetTask(
	repo out.Repository,
	cache out.Cache,
	producer out.EventProducer,
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
