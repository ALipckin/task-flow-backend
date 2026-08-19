package adapters

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/persistence"
	"tasks/internal/infrastructure/sharding/shard"
	"tasks/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements ports.Repository using pgxpool shards via shard.Manager.
type PostgresRepository struct {
	ShardManager *shard.ShardManager
	Router       out.ShardRouter
	shardIndex   out.TaskShardIndex
	log          *logger.Logger
}

func NewPostgresRepository(
	sm *shard.ShardManager,
	router out.ShardRouter,
	shardIndex out.TaskShardIndex,
	log *logger.Logger,
) *PostgresRepository {
	return &PostgresRepository{
		ShardManager: sm,
		Router:       router,
		shardIndex:   shardIndex,
		log:          log,
	}
}

func (r *PostgresRepository) Save(ctx context.Context, t domain.Task) error {
	shardIndex := r.Router.Resolve(t.PerformerId)
	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return errors.New("shard not found")
	}

	model := persistence.TaskFromDomain(t)

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO tasks (id, title, description, performer_id, creator_id, status, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL)
	`, model.ID, model.Title, model.Description, model.PerformerId, model.CreatorId, model.Status, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return err
	}

	for _, obs := range model.Observers {
		if _, err := tx.Exec(ctx, `
			INSERT INTO observers (user_id, task_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, NOW(), NOW(), NULL)
		`, obs.UserId, model.ID); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	_ = r.shardIndex.Set(ctx, t.ID, shardIndex)
	return nil
}

func (r *PostgresRepository) Find(ctx context.Context, filter out.TaskFilter) ([]domain.Task, error) {
	var all []domain.Task
	for _, db := range r.ShardManager.GetAllShards() {
		if db == nil {
			continue
		}
		tasks, err := r.findOnShard(ctx, db, filter)
		if err != nil {
			continue
		}
		all = append(all, tasks...)
	}
	return all, nil
}

func (r *PostgresRepository) findOnShard(ctx context.Context, db *pgxpool.Pool, filter out.TaskFilter) ([]domain.Task, error) {
	query := `
		SELECT id, title, description, performer_id, creator_id, status, created_at, updated_at, deleted_at
		FROM tasks
		WHERE deleted_at IS NULL
	`
	args := make([]interface{}, 0, 3)
	if filter.Title != "" {
		args = append(args, filter.Title)
		query += fmt.Sprintf(" AND title = $%d", len(args))
	}
	if filter.CreatorID != 0 {
		args = append(args, filter.CreatorID)
		query += fmt.Sprintf(" AND creator_id = $%d", len(args))
	}
	if filter.PerformerID != 0 {
		args = append(args, filter.PerformerID)
		query += fmt.Sprintf(" AND performer_id = $%d", len(args))
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := make([]persistence.Task, 0)
	for rows.Next() {
		var m persistence.Task
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.PerformerId, &m.CreatorId, &m.Status, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.attachObservers(ctx, db, models); err != nil {
		return nil, err
	}

	result := make([]domain.Task, 0, len(models))
	for _, m := range models {
		result = append(result, persistence.TaskToDomain(m))
	}

	return result, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, task domain.Task) error {
	taskID := task.ID
	shardIndex, err := r.shardIndex.Get(ctx, taskID)
	if err != nil {
		if !errors.Is(err, out.ErrNotFound) {
			return err
		}

		shardIndex, err = r.findShardIndexByTaskID(ctx, taskID)
		if err != nil {
			return err
		}
		_ = r.shardIndex.Set(ctx, taskID, shardIndex)
	}
	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return errors.New("shard not found")
	}

	cmd, err := db.Exec(ctx, `
		UPDATE tasks
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, task.DeletedAt, task.UpdatedAt, taskID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return out.ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, taskID uint) (*domain.Task, error) {
	shardIndex, err := r.shardIndex.Get(ctx, taskID)
	if err != nil {
		if errors.Is(err, out.ErrNotFound) {
			for idx, db := range r.ShardManager.GetAllShards() {
				task, err := r.getTaskByIDOnShard(ctx, db, taskID)
				if err == nil {
					_ = r.shardIndex.Set(ctx, task.ID, idx)
					dt := persistence.TaskToDomain(*task)
					return &dt, nil
				}
				if errors.Is(err, out.ErrNotFound) {
					continue
				}
				return nil, err
			}
			return nil, out.ErrNotFound

		}
		return nil, err
	}

	db := r.ShardManager.GetShardByIndex(shardIndex)
	if db == nil {
		return nil, errors.New("shard not found")
	}

	task, err := r.getTaskByIDOnShard(ctx, db, taskID)
	if err != nil {
		return nil, err
	}

	dt := persistence.TaskToDomain(*task)
	return &dt, nil
}

func (r *PostgresRepository) findShardIndexByTaskID(ctx context.Context, taskID uint) (int, error) {
	for idx, db := range r.ShardManager.GetAllShards() {
		if db == nil {
			continue
		}

		var found int
		err := db.QueryRow(ctx, `
			SELECT 1
			FROM tasks
			WHERE id = $1 AND deleted_at IS NULL
		`, taskID).Scan(&found)
		if err == nil {
			return idx, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		return -1, err
	}

	return -1, out.ErrNotFound
}

func (r *PostgresRepository) Update(ctx context.Context, task domain.Task) error {
	taskID := task.ID
	model := persistence.TaskFromDomain(task)

	currentShardIndex, err := r.shardIndex.Get(ctx, taskID)
	var fromShard *pgxpool.Pool

	if err == nil {
		fromShard = r.ShardManager.GetShardByIndex(currentShardIndex)
		if fromShard != nil {
			if _, getErr := r.getTaskPersistenceByID(ctx, fromShard, taskID); getErr != nil {
				if errors.Is(getErr, out.ErrNotFound) {
					return out.ErrNotFound
				}
				return getErr
			}
		}
	} else {
		if errors.Is(err, out.ErrNotFound) {
			r.log.Warn(ctx, "shard index miss for update", logger.ZapUint("task_id", taskID))
		} else {
			r.log.Warn(ctx, "shard index error on update", logger.ZapError(err))
		}
	}

	if fromShard == nil {
		allShards := r.ShardManager.GetAllShards()
		found := false
		for idx, sh := range allShards {
			if sh == nil {
				continue
			}

			if _, getErr := r.getTaskPersistenceByID(ctx, sh, taskID); getErr != nil {
				if errors.Is(getErr, out.ErrNotFound) {
					continue
				}
				return getErr
			}

			fromShard = sh
			currentShardIndex = idx
			found = true
			break
		}
		if !found {
			return out.ErrNotFound
		}
	}

	newShardIndex := r.Router.Resolve(model.PerformerId)
	needMigrate := task.PerformerChanged() && newShardIndex != currentShardIndex

	if needMigrate {
		toShard := r.ShardManager.GetShardByIndex(newShardIndex)
		if toShard == nil {
			return errors.New("target shard not found")
		}

		if err := r.migrateTaskToShard(ctx, &model, fromShard, toShard, newShardIndex); err != nil {
			return err
		}
	} else {
		tx, err := fromShard.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if _, err := tx.Exec(ctx, `
			UPDATE tasks
			SET title = $1, description = $2, performer_id = $3, creator_id = $4, status = $5, updated_at = $6
			WHERE id = $7
		`, model.Title, model.Description, model.PerformerId, model.CreatorId, model.Status, model.UpdatedAt, model.ID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM observers WHERE task_id = $1`, model.ID); err != nil {
			return err
		}

		for _, obs := range model.Observers {
			if _, err := tx.Exec(ctx, `
				INSERT INTO observers (user_id, task_id, created_at, updated_at, deleted_at)
				VALUES ($1, $2, NOW(), NOW(), NULL)
			`, obs.UserId, model.ID); err != nil {
				return err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) migrateTaskToShard(
	ctx context.Context,
	task *persistence.Task,
	fromShard *pgxpool.Pool,
	toShard *pgxpool.Pool,
	toIndex int,
) error {
	toTx, err := toShard.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = toTx.Rollback(ctx) }()

	if _, err := toTx.Exec(ctx, `
		INSERT INTO tasks (id, title, description, performer_id, creator_id, status, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NULL)
	`, task.ID, task.Title, task.Description, task.PerformerId, task.CreatorId, task.Status, task.CreatedAt); err != nil {
		return err
	}

	for _, obs := range task.Observers {
		if _, err := toTx.Exec(ctx, `
			INSERT INTO observers (user_id, task_id, created_at, updated_at, deleted_at)
			VALUES ($1, $2, NOW(), NOW(), NULL)
		`, obs.UserId, task.ID); err != nil {
			return err
		}
	}

	if err := toTx.Commit(ctx); err != nil {
		return err
	}

	fromTx, err := fromShard.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = fromTx.Rollback(ctx) }()

	if _, err := fromTx.Exec(ctx, `DELETE FROM observers WHERE task_id = $1`, task.ID); err != nil {
		return err
	}
	if _, err := fromTx.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, task.ID); err != nil {
		return err
	}

	if err := fromTx.Commit(ctx); err != nil {
		return err
	}

	if err := r.shardIndex.Set(ctx, task.ID, toIndex); err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) getTaskByIDOnShard(ctx context.Context, db *pgxpool.Pool, taskID uint) (*persistence.Task, error) {
	t, err := r.getTaskPersistenceByID(ctx, db, taskID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *PostgresRepository) getTaskPersistenceByID(ctx context.Context, db *pgxpool.Pool, taskID uint) (*persistence.Task, error) {
	var task persistence.Task
	err := db.QueryRow(ctx, `
		SELECT id, title, description, performer_id, creator_id, status, created_at, updated_at, deleted_at
		FROM tasks
		WHERE id = $1 AND deleted_at IS NULL
	`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.PerformerId, &task.CreatorId, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, out.ErrNotFound
		}
		return nil, err
	}

	obs, err := loadObservers(ctx, db, task.ID)
	if err != nil {
		return nil, err
	}
	task.Observers = obs
	return &task, nil
}

func (r *PostgresRepository) attachObservers(ctx context.Context, db *pgxpool.Pool, tasks []persistence.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	ids := make([]string, 0, len(tasks))
	args := make([]interface{}, 0, len(tasks))
	for i, t := range tasks {
		ids = append(ids, fmt.Sprintf("$%d", i+1))
		args = append(args, t.ID)
	}

	query := `
		SELECT id, user_id, task_id, created_at, updated_at, deleted_at
		FROM observers
		WHERE deleted_at IS NULL AND task_id IN (` + strings.Join(ids, ",") + `)
	`
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	byTask := make(map[uint][]persistence.Observer, len(tasks))
	for rows.Next() {
		var o persistence.Observer
		if err := rows.Scan(&o.ID, &o.UserId, &o.TaskId, &o.CreatedAt, &o.UpdatedAt, &o.DeletedAt); err != nil {
			return err
		}
		byTask[o.TaskId] = append(byTask[o.TaskId], o)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i := range tasks {
		tasks[i].Observers = byTask[tasks[i].ID]
	}

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

	out := make([]persistence.Observer, 0)
	for rows.Next() {
		var o persistence.Observer
		if err := rows.Scan(&o.ID, &o.UserId, &o.TaskId, &o.CreatedAt, &o.UpdatedAt, &o.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}

	return out, rows.Err()
}
