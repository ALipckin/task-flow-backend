package models

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ID range size per shard (default is 1M).
// Shard 0: 1..rangeSize, shard 1: rangeSize+1..2*rangeSize, etc.
const defaultRangeSize uint64 = 1_000_000

type ShardManager struct {
	shards    []*gorm.DB
	rangeSize uint64
	mu        sync.RWMutex
	// round-robin when creating a task
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

	var rangeSize uint64 = defaultRangeSize
	if s := os.Getenv("RANGE_SIZE"); s != "" {
		if n, err := strconv.ParseUint(s, 10, 64); err == nil && n > 0 {
			rangeSize = n
		}
	}

	ShardMgr = &ShardManager{
		shards:    shards,
		rangeSize: rangeSize,
	}
	log.Printf("ShardManager initialized with %d shards, range size %d", len(shards), rangeSize)
}

// GetShard returns a shard by task ID (range allocation).
// Shard i stores IDs in the range [i*rangeSize+1, (i+1)*rangeSize].
func (sm *ShardManager) GetShard(taskID uint) *gorm.DB {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.shards) == 0 || taskID == 0 {
		return nil
	}

	idx := (uint64(taskID) - 1) / sm.rangeSize
	if idx >= uint64(len(sm.shards)) {
		return nil
	}
	return sm.shards[idx]
}

// GetShardByIndex returns a shard by index (0-based).
func (sm *ShardManager) GetShardByIndex(index int) *gorm.DB {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if index < 0 || index >= len(sm.shards) {
		return nil
	}
	return sm.shards[index]
}

// NextShardIndex returns the shard index for a new task (round-robin).
func (sm *ShardManager) NextShardIndex() int {
	n := sm.GetShardCount()
	if n == 0 {
		return 0
	}
	i := atomic.AddUint32(&sm.nextShardIndex, 1)
	return int(i-1) % n
}

// AllocNextID allocates the next ID on the shard with index shardIndex and returns it.
func (sm *ShardManager) AllocNextID(shardIndex int) (uint, error) {
	shard := sm.GetShardByIndex(shardIndex)
	if shard == nil {
		return 0, gorm.ErrRecordNotFound
	}
	return AllocNextID(shard)
}

// RangeSize returns the size of the ID range per shard.
func (sm *ShardManager) RangeSize() uint64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.rangeSize
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
		shards:    shards,
		rangeSize: defaultRangeSize,
	}
}
