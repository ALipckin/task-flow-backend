package adapters

import (
	"log"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/kafke"
)

// InitializeInfrastructure centralizes previous initializer calls.
// It returns the initialized ShardManager (shard.ShardMgr) for callers that need it.
// NOTE: existing initializer functions may call log.Fatal/panic on failure; this function
// preserves that behavior and also returns the shard manager for DI.
func InitializeInfrastructure() *shard.ShardManager {
	// Initialize shard manager (DB shards)
	shard.InitShardManager()

	// Initialize cache (redis)
	cache.InitRedisFromEnv()

	// Initialize kafka producer
	kafke.InitProducer()

	// run migrations/sync for shards if available
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
