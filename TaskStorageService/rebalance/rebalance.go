package rebalance

import (
	"TaskStorageService/helpers"
	"TaskStorageService/initializers"
	"TaskStorageService/models"
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

// Run performs background rebalancing: for each performer_id whose shard on the ring
// changed after adding a new shard, migrates their tasks to the new shard.
// Postgres is unaware; the app copies data and updates the mapping in Redis.
func Run(ctx context.Context) {
	if models.ShardMgr == nil {
		return
	}
	allShards := models.ShardMgr.GetAllShards()
	for currentShardIndex, shard := range allShards {
		migratePerformerIDsFromShard(ctx, shard, currentShardIndex, allShards)
	}
}

// migratePerformerIDsFromShard for each distinct performer_id on the shard checks whether
// the current shard matches the one given by the ring; if not, migrates all that
// performer's tasks to the ring-assigned shard.
func migratePerformerIDsFromShard(ctx context.Context, shard *gorm.DB, currentShardIndex int, allShards []*gorm.DB) {
	var performerIDs []uint
	err := shard.Model(&models.Task{}).Distinct("performer_id").Pluck("performer_id", &performerIDs).Error
	if err != nil {
		log.Printf("[rebalance] shard %d: list performer_ids: %v", currentShardIndex, err)
		return
	}

	for _, performerID := range performerIDs {
		newShardIndex := models.ShardMgr.GetShardByPerformerIDIndex(performerID)
		if newShardIndex == currentShardIndex {
			continue
		}
		// performer_id is now on a different shard according to the ring — migrate their tasks
		newShard := models.ShardMgr.GetShardByIndex(newShardIndex)
		if newShard == nil {
			continue
		}
		migrateTasksByPerformer(ctx, shard, newShard, performerID, currentShardIndex, newShardIndex)
	}
}

func migrateTasksByPerformer(ctx context.Context, fromShard, toShard *gorm.DB, performerID uint, fromIndex, toIndex int) {
	var tasks []models.Task
	err := fromShard.Where("performer_id = ?", performerID).Preload("Observers").Find(&tasks).Error
	if err != nil {
		log.Printf("[rebalance] performer_id %d: list tasks: %v", performerID, err)
		return
	}

	for _, task := range tasks {
		if err := migrateOneTask(ctx, fromShard, toShard, task, fromIndex, toIndex); err != nil {
			log.Printf("[rebalance] task %d: %v", task.ID, err)
		}
	}
}

func migrateOneTask(ctx context.Context, fromShard, toShard *gorm.DB, task models.Task, fromIndex, toIndex int) error {
	// Clone task to target shard (same ID)
	if err := toShard.Create(&task).Error; err != nil {
		return err
	}
	// Create observers on target shard without old IDs (avoid PK conflicts)
	for _, obs := range task.Observers {
		newObs := models.Observer{UserId: obs.UserId, TaskId: task.ID}
		if err := toShard.Create(&newObs).Error; err != nil {
			return err
		}
	}
	// Remove from old shard (hard delete, not soft)
	if err := fromShard.Unscoped().Where("task_id = ?", task.ID).Delete(&models.Observer{}).Error; err != nil {
		return err
	}
	if err := fromShard.Unscoped().Delete(&task).Error; err != nil {
		return err
	}
	// Update task_id -> shard mapping in Redis
	if err := helpers.SetTaskShard(ctx, task.ID, toIndex); err != nil {
		return err
	}
	// Invalidate task cache so next GetTask loads from the new shard
	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))
	return nil
}

// RunBackground runs rebalancing in the background (e.g. after adding a shard).
// Call after updating config (new DB_SHARD_URLS) and restart, or when adding a shard
// dynamically — then call RebuildRing() first, then RunBackground().
func RunBackground(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			Run(ctx)
		}
	}
}
