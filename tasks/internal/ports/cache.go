package ports

import (
	"context"
	"tasks/internal/domain"
)

// Cache is a domain-level cache interface used by use-cases.
// Implementations should store/retrieve domain-level Task objects.
type Cache interface {
	Set(ctx context.Context, task domain.Task) error
	Get(ctx context.Context, taskID uint) (domain.Task, error)
}
