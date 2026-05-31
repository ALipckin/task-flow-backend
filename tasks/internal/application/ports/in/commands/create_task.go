package commands

import (
	"context"
	"tasks/internal/domain"
)

type CreateTaskHandler interface {
	Execute(ctx context.Context, cmd CreateTaskCommand) (domain.Task, error)
}

type CreateTaskCommand struct {
	Title       string
	Description string
	PerformerID uint
	CreatorID   uint
	ObserverIDs []uint
}
