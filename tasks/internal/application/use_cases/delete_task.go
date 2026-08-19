package use_cases

import (
	"context"
	"errors"
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/out"
)

type DeleteTask struct {
	repo       out.Repository
	cache      out.Cache
	shardIndex out.TaskShardIndex
	producer   out.EventProducer
}

func NewDeleteTask(
	repo out.Repository,
	cache out.Cache,
	shardIndex out.TaskShardIndex,
	producer out.EventProducer,
) *DeleteTask {
	return &DeleteTask{
		repo:       repo,
		cache:      cache,
		shardIndex: shardIndex,
		producer:   producer,
	}
}

func (uc *DeleteTask) Execute(
	ctx context.Context,
	cmd commands.DeleteTaskCommand,
) (bool, error) {
	taskID := uint(cmd.ID)

	task, err := uc.repo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, out.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if err := task.MarkDeleted(); err != nil {
		return false, err
	}

	if err := uc.repo.Delete(ctx, *task); err != nil {
		return false, err
	}

	_ = uc.cache.DeleteTask(ctx, taskID)
	_ = uc.shardIndex.Delete(ctx, taskID)

	if err := publishTaskEvents(ctx, uc.producer, task); err != nil {
		return true, err
	}

	return true, nil
}
