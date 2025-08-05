package helpers

import (
	"TaskStorageService/initializers"
	"TaskStorageService/models"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func CacheKey(taskID uint) string {
	return fmt.Sprintf("task:%d", taskID)
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
