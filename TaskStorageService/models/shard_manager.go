package models

import (
	"hash/fnv"
	"log"
	"os"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ShardManager struct {
	shards []*gorm.DB
	mu     sync.RWMutex
}

var ShardMgr *ShardManager

func InitShardManager() {
	shardURLs := os.Getenv("DB_SHARD_URLS")
	if shardURLs == "" {
		log.Fatal("DB_SHARD_URLS not set")
	}

	urls := strings.Split(shardURLs, ",")
	shards := make([]*gorm.DB, 0, len(urls))

	for i, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to shard %d: %v", i, err)
		}

		shards = append(shards, db)
		log.Printf("Connected to shard %d", i)
	}

	if len(shards) == 0 {
		log.Fatal("No shards configured")
	}

	ShardMgr = &ShardManager{
		shards: shards,
	}
	log.Printf("ShardManager initialized with %d shards", len(shards))
}

func (sm *ShardManager) GetShard(taskID uint) *gorm.DB {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.shards) == 0 {
		return nil
	}

	hash := fnv.New32a()
	hash.Write([]byte{byte(taskID), byte(taskID >> 8), byte(taskID >> 16), byte(taskID >> 24)})
	shardIndex := hash.Sum32() % uint32(len(sm.shards))

	return sm.shards[shardIndex]
}

func (sm *ShardManager) GetAllShards() []*gorm.DB {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	shards := make([]*gorm.DB, len(sm.shards))
	copy(shards, sm.shards)
	return shards
}

func (sm *ShardManager) GetShardCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.shards)
}

func NewShardManagerForTesting(shards []*gorm.DB) *ShardManager {
	return &ShardManager{
		shards: shards,
	}
}
