package queries

import (
	"context"
	"tasks/internal/domain"
)

type GetTaskHandler interface {
	Execute(ctx context.Context, cmd GetTaskQuery) (domain.Task, error)
}

type GetTaskQuery struct {
	ID uint64
}
