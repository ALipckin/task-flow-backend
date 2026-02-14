package use_case

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/domain/shard"
	"tasks/internal/ports"
)

type GetTasks struct {
	repo      ports.Repository
	sharder   *shard.ShardManager
	allocator ports.IDAllocator
}

// NewGetTasks constructs GetTasks use-case with its dependencies.
func NewGetTasks(
	repo ports.Repository,
	sharder *shard.ShardManager,
	allocator ports.IDAllocator,
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
	filter := ports.TaskFilter{
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
