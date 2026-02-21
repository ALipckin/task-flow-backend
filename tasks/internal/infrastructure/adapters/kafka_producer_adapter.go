package adapters

import (
	"context"
	"encoding/json"
	"log"
	"tasks/internal/domain"
	"tasks/internal/infrastructure/kafke"
	"tasks/internal/infrastructure/persistence"
)

type KafkaProducerAdapter struct{}

func NewKafkaProducerAdapter() *KafkaProducerAdapter { return &KafkaProducerAdapter{} }

func (a *KafkaProducerAdapter) PublishCreated(ctx context.Context, task domain.Task) error {
	// prepare payload
	payload := map[string]interface{}{
		"event": "TaskCreated",
		"id":    task.ID,
		"title": task.Title,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// send asynchronously and protect against panics / nil receivers
	go func(data []byte) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic while sending kafka message: %v", r)
			}
		}()

		// send via kafke package
		if err := kafke.SendMessageToKafka(data); err != nil {
			log.Printf("kafka send error: %v", err)
		}
	}(b)

	return nil
}

func (a *KafkaProducerAdapter) PublishDeleted(ctx context.Context, task domain.Task) error {
	// prepare payload
	payload := map[string]interface{}{
		"event": "TaskDeleted",
		"id":    task.ID,
		"title": task.Title,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// send asynchronously and protect against panics / nil receivers
	go func(data []byte) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic while sending kafka message: %v", r)
			}
		}()

		// send via kafke package
		if err := kafke.SendMessageToKafka(data); err != nil {
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
		if err := kafke.SendMessageToKafka(data); err != nil {
			log.Printf("kafka send error: %v", err)
		}
	}(b)

	return nil
}

func observerIDs(observers []persistence.Observer) []uint {
	if len(observers) == 0 {
		return nil
	}

	ids := make([]uint, len(observers))
	for i := range observers {
		ids[i] = observers[i].UserId
	}
	return ids
}
