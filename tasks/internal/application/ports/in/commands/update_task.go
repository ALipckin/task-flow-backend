package commands

import (
	"context"
	"tasks/internal/domain"
)

type UpdateTaskHandler interface {
	Execute(ctx context.Context, cmd UpdateTaskCommand) (domain.Task, error)
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
