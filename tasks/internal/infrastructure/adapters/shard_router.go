package adapters

import (
	"tasks/internal/application/ports/out"
	"tasks/internal/infrastructure/sharding/shard"
)

type shardRouter struct {
	manager *shard.ShardManager
}

func NewShardRouter(manager *shard.ShardManager) out.ShardRouter {
	return &shardRouter{manager: manager}
}

func (r *shardRouter) Resolve(performerID uint) int {
	return r.manager.Resolve(performerID)
}
