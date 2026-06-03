package shard

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Number of virtual nodes per physical shard (100–1000 for even distribution).
const defaultVnodesPerShard = 256

type ShardManager struct {
	shards  []*pgxpool.Pool
	ring    *consistentRing
	vnodes  int
	mu      sync.RWMutex
	nextIdx uint32
}

// NewShardManager builds the shard manager from provided shard URLs and vnodes.
func NewShardManager(shardURLs []string, vnodes int) *ShardManager {
	urls := normalizeShardURLs(shardURLs)
	if len(urls) == 0 {
		log.Fatal("DB_SHARD_URLS not set")
	}

	shards := make([]*pgxpool.Pool, 0, len(urls))
	for i, url := range urls {
		pool, err := pgxpool.New(context.Background(), url)
		if err != nil {
			log.Fatalf("Failed to create shard pool %d: %v", i, err)
		}
		if err := pool.Ping(context.Background()); err != nil {
			pool.Close()
			log.Fatalf("Failed to connect to shard %d: %v", i, err)
		}

		shards = append(shards, pool)
		log.Printf("Connected to shard %d", i)
	}

	if len(shards) == 0 {
		log.Fatal("No shards configured")
	}

	vnodes = normalizeVNodes(vnodes)
	ring := newConsistentRing(len(shards), vnodes)

	sm := &ShardManager{
		shards: shards,
		ring:   ring,
		vnodes: vnodes,
	}
	log.Printf("ShardManager initialized with %d shards, %d vnodes/shard (consistent ring)", len(shards), vnodes)
	return sm
}

// GetShardByPerformerID returns the shard for performer_id (ring key: performer:{id}).
// When performerID == 0, uses round-robin.
func (sm *ShardManager) GetShardByPerformerID(performerID uint) *pgxpool.Pool {
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
		i := atomic.AddUint32(&sm.nextIdx, 1)
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
func (sm *ShardManager) GetShardByIndex(index int) *pgxpool.Pool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if index < 0 || index >= len(sm.shards) {
		return nil
	}
	return sm.shards[index]
}

func (sm *ShardManager) GetAllShards() []*pgxpool.Pool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	shards := make([]*pgxpool.Pool, len(sm.shards))
	copy(shards, sm.shards)
	return shards
}

func (sm *ShardManager) GetShardCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.shards)
}

// Resolve is a convenience wrapper that returns the shard index for a performer ID.
// It keeps existing call sites that expect a Resolve method.
func (sm *ShardManager) Resolve(performerID uint) int {
	return sm.GetShardByPerformerIDIndex(performerID)
}

// RebuildRing rebuilds the ring with the current number of shards.
func (sm *ShardManager) RebuildRing() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	n := len(sm.shards)
	sm.ring.Rebuild(n, sm.vnodes)
	log.Printf("Ring rebuilt with %d shards, %d vnodes/shard", n, sm.vnodes)
}

func (sm *ShardManager) Close() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, shard := range sm.shards {
		if shard != nil {
			shard.Close()
		}
	}
}

func NewShardManagerForTesting(shards []*pgxpool.Pool) *ShardManager {
	ring := newConsistentRing(len(shards), defaultVnodesPerShard)
	return &ShardManager{
		shards: shards,
		ring:   ring,
		vnodes: defaultVnodesPerShard,
	}
}

func normalizeShardURLs(urls []string) []string {
	result := make([]string, 0, len(urls))
	for _, url := range urls {
		item := strings.TrimSpace(url)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}

func normalizeVNodes(vnodes int) int {
	if vnodes <= 0 {
		return defaultVnodesPerShard
	}
	return vnodes
}
