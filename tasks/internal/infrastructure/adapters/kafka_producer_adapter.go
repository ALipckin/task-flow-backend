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
	payload := map[string]interface{}{
		"event": "TaskCreated",
		"id":    task.ID,
		"title": task.Title,
	}
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

func (a *KafkaProducerAdapter) PublishDeleted(ctx context.Context, task domain.Task) error {
	payload := map[string]interface{}{
		"event": "TaskDeleted",
		"id":    task.ID,
		"title": task.Title,
	}
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

func (a *KafkaProducerAdapter) PublishUpdated(ctx context.Context, task domain.Task) error {
	message := map[string]interface{}{
		"event":         "TaskUpdated",
		"task_id":       task.ID,
		"title":         task.Title,
		"description":   task.Description,
		"performer_id":  task.PerformerId,
		"creator_id":    task.CreatorId,
		"observers_ids": observerIDs(task.Observers),
		"status":        task.Status,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}
	b, err := json.Marshal(message)
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
