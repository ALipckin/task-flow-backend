package ports

import (
	"context"
	"tasks/internal/domain"
)

// Repository represents persistence operations required by use-cases.
// Save stores a domain.Task into the given shard index (application chooses shard DB instance).
type Repository interface {
	Save(ctx context.Context, task domain.Task, shardIndex int) error
}

// IDAllocator generates IDs for new tasks.
// NextID should return the next unique ID (e.g., from a per-shard allocator).
type IDAllocator interface {
	NextID(ctx context.Context) (uint, error)
}
