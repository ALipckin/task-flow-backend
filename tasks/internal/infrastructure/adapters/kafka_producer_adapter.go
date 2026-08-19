package adapters

import (
	"context"
	"encoding/json"
	"log"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/kafke"
)

type KafkaProducerAdapter struct {
	producer *kafke.Producer
}

func NewKafkaProducerAdapter(producer *kafke.Producer) *KafkaProducerAdapter {
	return &KafkaProducerAdapter{producer: producer}
}

func (a *KafkaProducerAdapter) PublishCreated(ctx context.Context, task domain.Task) error {
	return a.publish(taskEventPayload("TaskCreated", task))
}

func (a *KafkaProducerAdapter) PublishDeleted(ctx context.Context, task domain.Task) error {
	return a.publish(taskEventPayload("TaskDeleted", task))
}

func (a *KafkaProducerAdapter) PublishUpdated(ctx context.Context, task domain.Task) error {
	return a.publish(taskEventPayload("TaskUpdated", task))
}

func (a *KafkaProducerAdapter) publish(payload map[string]interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	go func(data []byte) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic while sending kafka message: %v", r)
			}
		}()

		if err := a.producer.SendMessageToKafka(data); err != nil {
			log.Printf("kafka send error: %v", err)
		}
	}(b)

	return nil
}

func taskEventPayload(event string, task domain.Task) map[string]interface{} {
	return map[string]interface{}{
		"event":         event,
		"id":            task.ID,
		"task_id":       task.ID,
		"title":         task.Title,
		"description":   task.Description,
		"performer_id":  task.PerformerId,
		"creator_id":    task.CreatorId,
		"observers_ids": observerIDs(task.Observers),
		"status":        task.Status.String(),
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}
}

func observerIDs(observers []domain.Observer) []uint {
	if len(observers) == 0 {
		return nil
	}

	ids := make([]uint, len(observers))
	for i := range observers {
		ids[i] = observers[i].UserID
	}
	return ids
}
