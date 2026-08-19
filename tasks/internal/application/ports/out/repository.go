package out

import (
	"context"
	"tasks/internal/domain"
)

// TaskFilter represents query criteria for listing/searching tasks.
// Kept minimal for current needs (title, creator, performer).
type TaskFilter struct {
	Title       string
	CreatorID   uint
	PerformerID uint
}

// Repository represents persistence operations required by use-cases.
type Repository interface {
	Save(ctx context.Context, task domain.Task) error
	Find(ctx context.Context, filter TaskFilter) ([]domain.Task, error)
	Delete(ctx context.Context, task domain.Task) error
	GetByID(ctx context.Context, taskID uint) (*domain.Task, error)
	Update(ctx context.Context, task domain.Task) error
}

// IDAllocator generates IDs for new tasks.
// NextID should return the next unique ID (e.g., from a per-shard allocator).
type IDAllocator interface {
	NextID(ctx context.Context) (uint, error)
}
