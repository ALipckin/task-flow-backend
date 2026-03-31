package shard

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdAllocator stores the next ID for the shard (range allocation).
// In each shard, there is one row (id=1): next_id = start of range + number of IDs issued.
type IdAllocator struct {
	ID     int
	NextID int64
}

// SeedIdAllocator sets the initial value of next_id for the shard with index shardIndex.
// Range of shard i: [i*rangeSize+1, (i+1)*rangeSize].
func SeedIdAllocator(ctx context.Context, db *pgxpool.Pool, shardIndex int, rangeSize uint64) error {
	start := int64(shardIndex)*int64(rangeSize) + 1
	_, err := db.Exec(ctx, `
		INSERT INTO id_allocator (id, next_id)
		VALUES (1, $1)
		ON CONFLICT (id) DO NOTHING
	`, start)
	return err
}

// AllocNextID atomically increments next_id on the shard and returns the issued ID.
func AllocNextID(ctx context.Context, db *pgxpool.Pool) (uint, error) {
	var nextID int64
	err := db.QueryRow(ctx, `
		UPDATE id_allocator
		SET next_id = next_id + 1
		WHERE id = 1
		RETURNING next_id
	`).Scan(&nextID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, pgx.ErrNoRows
		}
		return 0, err
	}
	return uint(nextID), nil
}
