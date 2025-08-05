package helpers

import (
	"TaskStorageService/initializers"
	"TaskStorageService/models"
	"encoding/json"
)

func SendTaskEventToKafka(event string, task models.Task) error {
	message := map[string]interface{}{
		"event":         event,
		"task_id":       task.ID,
		"title":         task.Title,
		"description":   task.Description,
		"performer_id":  task.PerformerId,
		"creator_id":    task.CreatorId,
		"observers_ids": task.ObserverIDs(),
		"status":        task.Status,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return initializers.SendMessageToKafka(data)
}
