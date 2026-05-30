package queries

import (
	"context"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/sharding/shard"
)

type GetTasks struct {
	repo      out.Repository
	sharder   *shard.ShardManager
	allocator out.IDAllocator
}

// NewGetTasks constructs GetTasks use-case with its dependencies.
func NewGetTasks(
	repo out.Repository,
	sharder *shard.ShardManager,
	allocator out.IDAllocator,
) *GetTasks {
	return &GetTasks{
		repo:      repo,
		sharder:   sharder,
		allocator: allocator,
	}
}

type GetTasksCommand struct {
	Title       string
	PerformerID uint
	CreatorID   uint
}

// Execute returns tasks matching the command filter across all shards.
func (uc *GetTasks) Execute(ctx context.Context, cmd GetTasksCommand) ([]domain.Task, error) {
	filter := out.TaskFilter{
		Title:       cmd.Title,
		CreatorID:   cmd.CreatorID,
		PerformerID: cmd.PerformerID,
	}

	shardCount := uc.sharder.GetShardCount()
	var all []domain.Task
	for i := 0; i < shardCount; i++ {
		tasks, err := uc.repo.Find(ctx, filter, i)
		if err != nil {
			// skip shards that return an error (e.g., connection issues) but continue scanning others
			continue
		}
		all = append(all, tasks...)
	}

	return all, nil
}
