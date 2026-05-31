package queries

import (
	"context"
	"tasks/internal/domain"
)

type GetTasksHandler interface {
	Execute(ctx context.Context, cmd GetTasksQuery) ([]domain.Task, error)
}

type GetTasksQuery struct {
	Title       string
	PerformerID uint
	CreatorID   uint
}
