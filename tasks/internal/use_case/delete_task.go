package use_case

import (
	"context"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/ports"
)

type DeleteTask struct {
	repo     ports.Repository
	cache    ports.Cache
	producer ports.EventProducer
}

// NewDeleteTask constructs DeleteTask use-case with its dependencies.
func NewDeleteTask(
	repo ports.Repository,
	cache ports.Cache,
	producer ports.EventProducer,
) *DeleteTask {
	return &DeleteTask{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

type DeleteTaskCommand struct {
	ID uint64
}

func (uc *DeleteTask) Execute(
	ctx context.Context,
	cmd DeleteTaskCommand,
) (bool, error) {
	taskID := uint(cmd.ID)

	task, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		return false, nil
	}

	if err := uc.repo.Delete(ctx, taskID); err != nil {
		return false, err
	}

	_ = cache.DeleteTaskCache(ctx, taskID)
	_ = cache.DelTaskShard(ctx, taskID)
	// If we have err here, we return true, because task is already deleted, but log the error for debugging
	if err := uc.producer.PublishDeleted(ctx, *task); err != nil {
		return true, err
	}

	return true, nil
}
