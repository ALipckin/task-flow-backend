package out

// ShardRouter resolves which shard owns data for a performer.
// Routing stays in infrastructure; use-cases must not pass shard indices.
type ShardRouter interface {
	Resolve(performerID uint) int
}
