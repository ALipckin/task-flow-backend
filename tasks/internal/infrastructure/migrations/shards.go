package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"tasks/internal/infrastructure/sharding/shard"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

// SyncDatabaseForShards applies SQL migrations on every shard.
// Add a new migration by creating a new file in sql/ (e.g. sql/0002_add_column.sql).
func SyncDatabaseForShards(ctx context.Context, shardMgr *shard.ShardManager) error {
	if shardMgr == nil {
		return fmt.Errorf("shard manager is nil")
	}

	allShards := shardMgr.GetAllShards()
	if len(allShards) == 0 {
		return fmt.Errorf("no shards configured")
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	var firstErr error
	var failed int

	for i, db := range allShards {
		if err := applyMigrationsToShard(ctx, db, migrations); err != nil {
			log.Printf("Error migrating shard %d: %v", i, err)
			if firstErr == nil {
				firstErr = err
			}
			failed++
			continue
		}
		log.Printf("Successful shard migration %d", i)
	}

	if failed > 0 {
		return fmt.Errorf("%d shard migrations failed: %w", failed, firstErr)
	}

	log.Println("all migration successful")
	return nil
}

type migration struct {
	version string
	name    string
	sql     string
}

func loadMigrations() ([]migration, error) {
	names, err := fs.Glob(migrationFiles, "sql/*.sql")
	if err != nil {
		return nil, fmt.Errorf("glob migrations: %w", err)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no migration files found")
	}

	sort.Strings(names)

	out := make([]migration, 0, len(names))
	for _, n := range names {
		b, err := migrationFiles.ReadFile(n)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", n, err)
		}
		base := filepath.Base(n)
		version := strings.TrimSuffix(base, filepath.Ext(base))
		out = append(out, migration{
			version: version,
			name:    base,
			sql:     string(b),
		})
	}
	return out, nil
}

func applyMigrationsToShard(ctx context.Context, db *pgxpool.Pool, migrations []migration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	for _, m := range migrations {
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, m.version).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		tx, err := db.Begin(ctx)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, m.sql); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", m.name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.version); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
