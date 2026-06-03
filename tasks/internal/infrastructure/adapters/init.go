package adapters

import (
	"context"
	"log"
	"tasks/internal/config"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/kafke"
	"tasks/internal/infrastructure/migrations"
	"tasks/internal/infrastructure/sharding/shard"
)

// Infrastructure holds shared infrastructure dependencies.
type Infrastructure struct {
	ShardManager *shard.ShardManager
	Redis        *cache.Store
	Kafka        *kafke.Producer
}

// NewInfrastructure wires shard manager, Redis, and Kafka from config.
func NewInfrastructure(cfg *config.Config) *Infrastructure {
	if cfg == nil {
		log.Fatalf("config is nil")
	}

	sm := shard.NewShardManager(cfg.DBShardURLs, cfg.VNodesPerShard)
	redis := cache.NewStore(cfg.RedisURL)
	kafka := kafke.NewProducer(cfg.KafkaBrokers)

	if err := migrations.SyncDatabaseForShards(context.Background(), sm); err != nil {
		log.Fatalf("failed to migrate shards: %v", err)
	}

	return &Infrastructure{
		ShardManager: sm,
		Redis:        redis,
		Kafka:        kafka,
	}
}

// Close releases infrastructure resources.
func (i *Infrastructure) Close() {
	if i == nil {
		return
	}
	if i.Kafka != nil {
		if err := i.Kafka.Close(); err != nil {
			log.Printf("Error closing kafka producer: %v", err)
		}
	}
	if i.Redis != nil {
		if err := i.Redis.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}
	if i.ShardManager != nil {
		i.ShardManager.Close()
	}
}
