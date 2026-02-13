package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"tasks/internal/infrastructure/persistence"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

const (
	taskIDCounterKey = "task:id_counter"
	taskShardKeyFmt  = "task:shard:%d"
)

func InitRedisFromEnv() {
	url := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(opts)
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}
}

func CloseRedis() error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Close()
}

// IsNilError reports whether err indicates a missing key in Redis (redis.Nil).
func IsNilError(err error) bool {
	return errors.Is(err, redis.Nil)
}

func CacheKey(taskID uint) string {
	return fmt.Sprintf("task:%d", taskID)
}

// AllocTaskID returns the next global task ID (Redis INCR).
func AllocTaskID(ctx context.Context) (uint, error) {
	n, err := redisClient.Incr(ctx, taskIDCounterKey).Result()
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// SetTaskShard stores the task_id -> shard_index mapping (for GetTask/Update/Delete).
func SetTaskShard(ctx context.Context, taskID uint, shardIndex int) error {
	return redisClient.Set(ctx, fmt.Sprintf(taskShardKeyFmt, taskID), shardIndex, 0).Err()
}

// GetTaskShard returns the shard index where the task lives. Returns error if not found.
func GetTaskShard(ctx context.Context, taskID uint) (int, error) {
	s, err := redisClient.Get(ctx, fmt.Sprintf(taskShardKeyFmt, taskID)).Result()
	if err != nil {
		return -1, err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}
	return i, nil
}

// DelTaskShard removes the mapping (on task delete or after migration update).
func DelTaskShard(ctx context.Context, taskID uint) error {
	return redisClient.Del(ctx, fmt.Sprintf(taskShardKeyFmt, taskID)).Err()
}

func CacheSetTask(ctx context.Context, task persistence.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, CacheKey(task.ID), data, 10*time.Minute).Err()
}

func CacheGetTask(ctx context.Context, taskID uint) (*persistence.Task, error) {
	data, err := redisClient.Get(ctx, CacheKey(taskID)).Result()
	if err != nil {
		return nil, err
	}
	var task persistence.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// DeleteTaskCache removes the cached serialized task by key.
func DeleteTaskCache(ctx context.Context, taskID uint) error {
	return redisClient.Del(ctx, CacheKey(taskID)).Err()
}
