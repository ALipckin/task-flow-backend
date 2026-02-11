package models

import (
	"fmt"
	"sort"
	"sync"

	"github.com/cespare/xxhash/v2"
)

// Hash space of the ring: 0 .. 2^32-1.
const hashSpace = uint64(1 << 32)

// ringNode is a point on the ring: vnode hash and physical shard index.
type ringNode struct {
	hash  uint32
	shard int
}

// consistentRing is a ring with virtual nodes.
// Sorted by hash ascending; a key belongs to the first shard clockwise.
type consistentRing struct {
	nodes []ringNode // sorted by hash
	mu    sync.RWMutex
}

// hash32 returns a hash in the space 0 .. 2^32-1 (lower 32 bits of xxhash).
func hash32(data []byte) uint32 {
	return uint32(xxhash.Sum64(data) % hashSpace)
}

// newConsistentRing builds the ring: for each shard adds vnodesPerShard virtual nodes.
// Vnode identifier: "shard-{i}-vnode-{j}"; hash and place on the ring.
func newConsistentRing(numShards, vnodesPerShard int) *consistentRing {
	if numShards <= 0 {
		return &consistentRing{nodes: nil}
	}
	nodes := make([]ringNode, 0, numShards*vnodesPerShard)
	for i := 0; i < numShards; i++ {
		for j := 0; j < vnodesPerShard; j++ {
			key := []byte(fmt.Sprintf("shard-%d-vnode-%d", i, j))
			h := hash32(key)
			nodes = append(nodes, ringNode{hash: h, shard: i})
		}
	}
	sort.Slice(nodes, func(a, b int) bool { return nodes[a].hash < nodes[b].hash })
	return &consistentRing{nodes: nodes}
}

// GetShard returns the shard index for the key: first shard clockwise (lower_bound).
// If the key is greater than all points, wrap to the first (nodes[0]).
func (r *consistentRing) GetShard(key []byte) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.nodes) == 0 {
		return 0
	}
	h := hash32(key)
	// lower_bound: first i such that nodes[i].hash >= h
	i := sort.Search(len(r.nodes), func(i int) bool {
		return r.nodes[i].hash >= h
	})
	if i == len(r.nodes) {
		i = 0
	}
	return r.nodes[i].shard
}

// Rebuild rebuilds the ring with a new number of shards (when adding a shard).
func (r *consistentRing) Rebuild(numShards, vnodesPerShard int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if numShards <= 0 {
		r.nodes = nil
		return
	}
	nodes := make([]ringNode, 0, numShards*vnodesPerShard)
	for i := 0; i < numShards; i++ {
		for j := 0; j < vnodesPerShard; j++ {
			key := []byte(fmt.Sprintf("shard-%d-vnode-%d", i, j))
			nodes = append(nodes, ringNode{hash: hash32(key), shard: i})
		}
	}
	sort.Slice(nodes, func(a, b int) bool { return nodes[a].hash < nodes[b].hash })
	r.nodes = nodes
}
