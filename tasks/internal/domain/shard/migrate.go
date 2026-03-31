package shard

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncDatabaseForShards creates required tables/indexes in every shard.
func SyncDatabaseForShards() {
	if ShardMgr == nil {
		log.Fatal("ShardManager not initialized")
	}

	allShards := ShardMgr.GetAllShards()

	for i, db := range allShards {
		if err := migrateShard(db); err != nil {
			log.Printf("Error migrating shard %d: %v", i, err)
			continue
		}
		log.Printf("Successful shard migration %d", i)
	}
	log.Println("all migration successful")
}

func migrateShard(db *pgxpool.Pool) error {
	ctx := context.Background()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id BIGINT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			performer_id BIGINT NOT NULL,
			creator_id BIGINT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_performer_id ON tasks (performer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_creator_id ON tasks (creator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks (deleted_at)`,
		`CREATE TABLE IF NOT EXISTS observers (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_observers_user_id ON observers (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_observers_task_id ON observers (task_id)`,
		`CREATE TABLE IF NOT EXISTS id_allocator (
			id INT PRIMARY KEY,
			next_id BIGINT NOT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(ctx, q); err != nil {
			return err
		}
	}

	return nil
}
