package use_cases

import (
	"context"
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

type UpdateTask struct {
	repo     out.Repository
	cache    out.Cache
	producer out.EventProducer
}

func NewUpdateTask(
	repo out.Repository,
	cache out.Cache,
	producer out.EventProducer,
) *UpdateTask {
	return &UpdateTask{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

func (uc *UpdateTask) Execute(ctx context.Context, cmd commands.UpdateTaskCommand) (domain.Task, error) {
	task, err := uc.repo.GetByID(ctx, uint(cmd.ID))
	if err != nil {
		return domain.Task{}, err
	}

	if err := task.ApplyUpdate(
		cmd.Title,
		cmd.Description,
		cmd.Status,
		cmd.PerformerID,
		cmd.CreatorID,
		uint64SliceToUint(cmd.ObserverIDs),
	); err != nil {
		return domain.Task{}, err
	}

	if err := uc.repo.Update(ctx, *task); err != nil {
		return domain.Task{}, err
	}

	_ = uc.cache.DeleteTask(ctx, task.ID)
	_ = publishTaskEvents(ctx, uc.producer, task)

	return *task, nil
}

func uint64SliceToUint(src []uint64) []uint {
	if len(src) == 0 {
		return nil
	}

	dst := make([]uint, len(src))
	for i := range src {
		dst[i] = uint(src[i])
	}
	return dst
}
