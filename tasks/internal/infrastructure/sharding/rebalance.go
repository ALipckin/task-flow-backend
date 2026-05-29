package sharding

import (
	"context"
	"log"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/cache"
	"tasks/internal/infrastructure/persistence"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run performs background rebalancing: for each performer_id whose shard on the ring
// changed after adding a new shard, migrates their tasks to the new shard.
// Postgres is unaware; the app copies data and updates the mapping in Redis.
func Run(ctx context.Context) {
	if shard.ShardMgr == nil {
		return
	}
	allShards := shard.ShardMgr.GetAllShards()
	for currentShardIndex, sh := range allShards {
		migratePerformerIDsFromShard(ctx, sh, currentShardIndex)
	}
}

func migratePerformerIDsFromShard(ctx context.Context, sh *pgxpool.Pool, currentShardIndex int) {
	rows, err := sh.Query(ctx, `
		SELECT DISTINCT performer_id
		FROM tasks
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		log.Printf("[rebalance] shard %d: list performer_ids: %v", currentShardIndex, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var performerID uint
		if err := rows.Scan(&performerID); err != nil {
			log.Printf("[rebalance] shard %d: scan performer_id: %v", currentShardIndex, err)
			continue
		}

		newShardIndex := shard.ShardMgr.GetShardByPerformerIDIndex(performerID)
		if newShardIndex == currentShardIndex {
			continue
		}

		newShard := shard.ShardMgr.GetShardByIndex(newShardIndex)
		if newShard == nil {
			continue
		}
		migrateTasksByPerformer(ctx, sh, newShard, performerID, newShardIndex)
	}
}

func migrateTasksByPerformer(ctx context.Context, fromShard, toShard *pgxpool.Pool, performerID uint, toIndex int) {
	rows, err := fromShard.Query(ctx, `
		SELECT id, title, description, performer_id, creator_id, status, created_at, updated_at, deleted_at
		FROM tasks
		WHERE performer_id = $1 AND deleted_at IS NULL
	`, performerID)
	if err != nil {
		log.Printf("[rebalance] performer_id %d: list tasks: %v", performerID, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task persistence.Task
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.PerformerId,
			&task.CreatorId,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeletedAt,
		); err != nil {
			log.Printf("[rebalance] performer_id %d: scan task: %v", performerID, err)
			continue
		}

		obs, err := loadObservers(ctx, fromShard, task.ID)
		if err != nil {
			log.Printf("[rebalance] task %d: load observers: %v", task.ID, err)
			continue
		}
		task.Observers = obs

		if err := migrateOneTask(ctx, fromShard, toShard, task, toIndex); err != nil {
			log.Printf("[rebalance] task %d: %v", task.ID, err)
		}
	}
}

func migrateOneTask(ctx context.Context, fromShard, toShard *pgxpool.Pool, task persistence.Task, toIndex int) error {
	if _, err := toShard.Exec(ctx, `
		INSERT INTO tasks (id, title, description, performer_id, creator_id, status, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NULL)
	`, task.ID, task.Title, task.Description, task.PerformerId, task.CreatorId, task.Status, task.CreatedAt); err != nil {
		return err
	}

	for _, obs := range task.Observers {
		if _, err := toShard.Exec(ctx, `
			INSERT INTO observers (user_id, task_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, NOW(), NOW(), NULL)
		`, obs.UserId, task.ID); err != nil {
			return err
		}
	}

	if _, err := fromShard.Exec(ctx, `DELETE FROM observers WHERE task_id = $1`, task.ID); err != nil {
		return err
	}
	if _, err := fromShard.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, task.ID); err != nil {
		return err
	}

	if err := cache.SetTaskShard(ctx, task.ID, toIndex); err != nil {
		return err
	}
	_ = cache.DeleteTaskCache(ctx, task.ID)
	return nil
}

func loadObservers(ctx context.Context, db *pgxpool.Pool, taskID uint) ([]persistence.Observer, error) {
	rows, err := db.Query(ctx, `
		SELECT id, user_id, task_id, created_at, updated_at, deleted_at
		FROM observers
		WHERE task_id = $1 AND deleted_at IS NULL
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []persistence.Observer
	for rows.Next() {
		var o persistence.Observer
		if err := rows.Scan(&o.ID, &o.UserId, &o.TaskId, &o.CreatedAt, &o.UpdatedAt, &o.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// RunBackground runs rebalancing in the background (e.g. after adding a shard).
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
