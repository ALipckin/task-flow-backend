package models

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Number of virtual nodes per physical shard (100–1000 for even distribution).
const defaultVnodesPerShard = 256

type ShardManager struct {
	shards []*gorm.DB
	ring   *consistentRing
	mu     sync.RWMutex
	// round-robin when performer_id == 0 (fallback)
	nextShardIndex uint32
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

	vnodes := defaultVnodesPerShard
	if s := os.Getenv("VNODES_PER_SHARD"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			vnodes = n
		}
	}

	ring := newConsistentRing(len(shards), vnodes)

	ShardMgr = &ShardManager{
		shards: shards,
		ring:   ring,
	}
	log.Printf("ShardManager initialized with %d shards, %d vnodes/shard (consistent ring)", len(shards), vnodes)
}

// GetShardByPerformerID returns the shard for performer_id (ring key: performer:{id}).
// When performerID == 0, uses round-robin.
func (sm *ShardManager) GetShardByPerformerID(performerID uint) *gorm.DB {
	idx := sm.GetShardByPerformerIDIndex(performerID)
	return sm.GetShardByIndex(idx)
}

// GetShardByPerformerIDIndex returns the shard index for performer_id (first shard clockwise on the ring).
// When performerID == 0, uses round-robin.
func (sm *ShardManager) GetShardByPerformerIDIndex(performerID uint) int {
	sm.mu.RLock()
	n := len(sm.shards)
	sm.mu.RUnlock()
	if n == 0 {
		return 0
	}
	if performerID == 0 {
		i := atomic.AddUint32(&sm.nextShardIndex, 1)
		return int(i-1) % n
	}
	key := []byte(fmt.Sprintf("performer:%d", performerID))
	sm.mu.RLock()
	idx := sm.ring.GetShard(key)
	sm.mu.RUnlock()
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}

// GetShardByIndex returns the shard by index (0-based).
func (sm *ShardManager) GetShardByIndex(index int) *gorm.DB {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if index < 0 || index >= len(sm.shards) {
		return nil
	}
	return sm.shards[index]
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

// RebuildRing rebuilds the ring with the current number of shards (after adding a shard).
// On restart with new DB_SHARD_URLS, InitShardManager already builds a new ring.
func (sm *ShardManager) RebuildRing() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	n := len(sm.shards)
	vnodes := defaultVnodesPerShard
	if s := os.Getenv("VNODES_PER_SHARD"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			vnodes = v
		}
	}
	sm.ring.Rebuild(n, vnodes)
	log.Printf("Ring rebuilt with %d shards, %d vnodes/shard", n, vnodes)
}

func NewShardManagerForTesting(shards []*gorm.DB) *ShardManager {
	ring := newConsistentRing(len(shards), defaultVnodesPerShard)
	return &ShardManager{
		shards: shards,
		ring:   ring,
	}
}
