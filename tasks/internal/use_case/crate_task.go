package use_case

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/domain/shard"
	"tasks/internal/ports"
)

type CreateTask struct {
	repo      ports.Repository
	cache     ports.Cache
	producer  ports.EventProducer
	sharder   *shard.ShardManager
	allocator ports.IDAllocator
}

// NewCreateTask constructs CreateTask use-case with its dependencies.
func NewCreateTask(repo ports.Repository, cache ports.Cache, producer ports.EventProducer, sharder *shard.ShardManager, allocator ports.IDAllocator) *CreateTask {
	return &CreateTask{repo: repo, cache: cache, producer: producer, sharder: sharder, allocator: allocator}
}

type CreateTaskCommand struct {
	Title       string
	Description string
	PerformerID uint
	CreatorID   uint
	ObserverIDs []uint
}

func (uc *CreateTask) Execute(ctx context.Context, cmd CreateTaskCommand) (domain.Task, error) {

	id, err := uc.allocator.NextID(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	shardIndex := uc.sharder.Resolve(cmd.PerformerID)

	task := domain.NewTask(id, cmd.Title, cmd.Description, cmd.CreatorID, cmd.PerformerID)

	if err := uc.repo.Save(ctx, task, shardIndex); err != nil {
		return domain.Task{}, err
	}

	_ = uc.cache.Set(ctx, task)
	_ = uc.producer.PublishCreated(ctx, task)

	return task, nil
}
