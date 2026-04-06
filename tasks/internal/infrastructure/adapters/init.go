package adapters

import (
	"log"
	"tasks/internal/config"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/kafke"
)

// InitializeInfrastructure centralizes previous initializer calls.
// It returns the initialized ShardManager (shard.ShardMgr) for callers that need it.
func InitializeInfrastructure(cfg *config.Config) *shard.ShardManager {
	if cfg == nil {
		log.Fatalf("config is nil")
	}

	shard.InitShardManager(cfg.DBShardURLs, cfg.VNodesPerShard)
	cache.InitRedis(cfg.RedisURL)
	kafke.InitProducer(cfg.KafkaBrokers)
	shard.SyncDatabaseForShards()

	if shard.ShardMgr == nil {
		log.Fatalf("shard manager not initialized")
	}
	return shard.ShardMgr
}

// CleanupInfrastructure performs cleanup for infra packages initialized above.
func CleanupInfrastructure() {
	if err := kafke.CloseProducer(); err != nil {
		log.Printf("Error closing kafka producer: %v", err)
	}
	if err := cache.CloseRedis(); err != nil {
		log.Printf("Error closing redis client: %v", err)
	}
}
