package helpers

import (
	"TaskStorageService/initializers"
	"TaskStorageService/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const (
	taskIDCounterKey = "task:id_counter"
	taskShardKeyFmt  = "task:shard:%d"
)

func CacheKey(taskID uint) string {
	return fmt.Sprintf("task:%d", taskID)
}

// AllocTaskID returns the next global task ID (Redis INCR).
func AllocTaskID(ctx context.Context) (uint, error) {
	n, err := initializers.RedisClient.Incr(ctx, taskIDCounterKey).Result()
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// SetTaskShard stores the task_id -> shard_index mapping (for GetTask/Update/Delete).
func SetTaskShard(ctx context.Context, taskID uint, shardIndex int) error {
	return initializers.RedisClient.Set(ctx, fmt.Sprintf(taskShardKeyFmt, taskID), shardIndex, 0).Err()
}

// GetTaskShard returns the shard index where the task lives. Returns error if not found.
func GetTaskShard(ctx context.Context, taskID uint) (int, error) {
	s, err := initializers.RedisClient.Get(ctx, fmt.Sprintf(taskShardKeyFmt, taskID)).Result()
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
	return initializers.RedisClient.Del(ctx, fmt.Sprintf(taskShardKeyFmt, taskID)).Err()
}

func CacheSetTask(ctx context.Context, task models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return initializers.RedisClient.Set(ctx, CacheKey(task.ID), data, 10*time.Minute).Err()
}

func CacheGetTask(ctx context.Context, taskID uint) (*models.Task, error) {
	data, err := initializers.RedisClient.Get(ctx, CacheKey(taskID)).Result()
	if err != nil {
		return nil, err
	}
	var task models.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, err
	}
	return &task, nil
}
tasks