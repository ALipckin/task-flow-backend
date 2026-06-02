package out

import "context"

// TaskShardIndex stores task_id -> shard_index mappings used for shard routing.
type TaskShardIndex interface {
	Get(ctx context.Context, taskID uint) (shardIndex int, err error)
	Set(ctx context.Context, taskID uint, shardIndex int) error
	Delete(ctx context.Context, taskID uint) error
}
