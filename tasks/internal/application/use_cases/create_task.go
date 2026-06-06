package use_cases

import (
	"context"
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

type CreateTask struct {
	repo      out.Repository
	cache     out.Cache
	producer  out.EventProducer
	allocator out.IDAllocator
}

func NewCreateTask(
	repo out.Repository,
	cache out.Cache,
	producer out.EventProducer,
	allocator out.IDAllocator,
) *CreateTask {
	return &CreateTask{repo: repo, cache: cache, producer: producer, allocator: allocator}
}

func (uc *CreateTask) Execute(ctx context.Context, cmd commands.CreateTaskCommand) (domain.Task, error) {
	id, err := uc.allocator.NextID(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	task, err := domain.NewTask(
		id,
		cmd.Title,
		cmd.Description,
		cmd.CreatorID,
		cmd.PerformerID,
		cmd.ObserverIDs,
	)
	if err != nil {
		return domain.Task{}, err
	}

	if err := uc.repo.Save(ctx, task); err != nil {
		return domain.Task{}, err
	}

	_ = uc.cache.SetTask(ctx, task)
	_ = uc.producer.PublishCreated(ctx, task)
	task.PullEvents()

	return task, nil
}
