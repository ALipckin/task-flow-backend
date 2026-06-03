package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"tasks/internal/infrastructure/persistence"

	"github.com/redis/go-redis/v9"
)

const (
	taskIDCounterKey = "task:id_counter"
	taskShardKeyFmt  = "task:shard:%d"
)

// Store provides Redis-backed task cache and ID allocation.
type Store struct {
	client *redis.Client
}

// NewStore connects to Redis using the provided URL.
func NewStore(redisURL string) *Store {
	if redisURL == "" {
		panic("REDIS_URL not set")
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opts)

	const attempts = 10
	const delay = 2 * time.Second
	var lastErr error
	for i := 1; i <= attempts; i++ {
		if _, lastErr = client.Ping(context.Background()).Result(); lastErr == nil {
			return &Store{client: client}
		}
		if i < attempts {
			time.Sleep(delay)
		}
	}
	log.Fatalf(
		"redis unreachable at %s after %d attempts: %v (ensure redis is on task-network: docker compose up -d redis --force-recreate)",
		redisURL, attempts, lastErr,
	)
	return nil
}

func (s *Store) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

// IsNilError reports whether err indicates a missing key in Redis (redis.Nil).
func IsNilError(err error) bool {
	return errors.Is(err, redis.Nil)
}

func CacheKey(taskID uint) string {
	return fmt.Sprintf("task:%d", taskID)
}

func (s *Store) AllocTaskID(ctx context.Context) (uint, error) {
	n, err := s.client.Incr(ctx, taskIDCounterKey).Result()
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

func (s *Store) SetTaskShard(ctx context.Context, taskID uint, shardIndex int) error {
	return s.client.Set(ctx, fmt.Sprintf(taskShardKeyFmt, taskID), shardIndex, 0).Err()
}

func (s *Store) GetTaskShard(ctx context.Context, taskID uint) (int, error) {
	val, err := s.client.Get(ctx, fmt.Sprintf(taskShardKeyFmt, taskID)).Result()
	if err != nil {
		return -1, err
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func (s *Store) DelTaskShard(ctx context.Context, taskID uint) error {
	return s.client.Del(ctx, fmt.Sprintf(taskShardKeyFmt, taskID)).Err()
}

func (s *Store) SetTask(ctx context.Context, task persistence.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, CacheKey(task.ID), data, 10*time.Minute).Err()
}

func (s *Store) GetTask(ctx context.Context, taskID uint) (*persistence.Task, error) {
	data, err := s.client.Get(ctx, CacheKey(taskID)).Result()
	if err != nil {
		return nil, err
	}
	var task persistence.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) DeleteTaskCache(ctx context.Context, taskID uint) error {
	return s.client.Del(ctx, CacheKey(taskID)).Err()
}
