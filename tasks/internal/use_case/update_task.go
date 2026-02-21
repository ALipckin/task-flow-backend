package use_case

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/ports"
)

type UpdateTask struct {
	repo     ports.Repository
	producer ports.EventProducer
}

func NewUpdateTask(
	repo ports.Repository,
	producer ports.EventProducer,
) *UpdateTask {
	return &UpdateTask{
		repo:     repo,
		producer: producer,
	}
}

type UpdateTaskCommand struct {
	ID          uint64
	Title       string
	Description string
	Status      string
	PerformerID uint
	CreatorID   uint
	ObserverIDs []uint64
}

func (uc *UpdateTask) Execute(ctx context.Context, cmd UpdateTaskCommand) (domain.Task, error) {
	input := ports.UpdateTaskInput{
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
