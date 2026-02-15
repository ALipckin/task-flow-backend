package ports

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
// Save stores a domain.Task into the given shard index (application chooses shard DB instance).
type Repository interface {
	Save(ctx context.Context, task domain.Task, shardIndex int) error
	// Find returns tasks matching the filter from the specified shard index.
	// If shardIndex is negative, caller may interpret it as "search all shards" (adapter-specific).
	Find(ctx context.Context, filter TaskFilter, shardIndex int) ([]domain.Task, error)
	Delete(ctx context.Context, taskID uint) error
	GetByID(ctx context.Context, taskID uint) (*domain.Task, error)
}

// IDAllocator generates IDs for new tasks.
// NextID should return the next unique ID (e.g., from a per-shard allocator).
type IDAllocator interface {
	NextID(ctx context.Context) (uint, error)
}
