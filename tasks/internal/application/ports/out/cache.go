package out

import (
	"context"
	"tasks/internal/domain"
)

type Cache interface {
	SetTask(ctx context.Context, task domain.Task) error
	GetTask(ctx context.Context, taskID uint) (domain.Task, error)
	DeleteTask(ctx context.Context, taskID uint) error
}
