package use_cases

import (
	"context"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/sharding/shard"
)

type GetTasks struct {
	repo      out.Repository
	sharder   *shard.ShardManager
	allocator out.IDAllocator
}

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

func (uc *GetTasks) Execute(ctx context.Context, cmd queries.GetTasksQuery) ([]domain.Task, error) {
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
