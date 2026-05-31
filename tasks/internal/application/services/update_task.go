package services

import (
	"context"
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
)

type UpdateTask struct {
	repo     out.Repository
	producer out.EventProducer
}

func NewUpdateTask(
	repo out.Repository,
	producer out.EventProducer,
) *UpdateTask {
	return &UpdateTask{
		repo:     repo,
		producer: producer,
	}
}

func (uc *UpdateTask) Execute(ctx context.Context, cmd commands.UpdateTaskCommand) (domain.Task, error) {
	input := out.UpdateTaskInput{
		ID:          uint(cmd.ID),
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      cmd.Status,
		PerformerID: cmd.PerformerID,
		CreatorID:   cmd.CreatorID,
		ObserverIDs: uint64SliceToUint(cmd.ObserverIDs),
	}

	task, err := uc.repo.Update(ctx, input)
	if err != nil {
		return domain.Task{}, err
	}

	if uc.producer != nil {
		_ = uc.producer.PublishUpdated(ctx, *task)
	}

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
