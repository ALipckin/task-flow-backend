package kafka

import (
	"context"
	"encoding/json"
	"notification/internal/domain"
	"notification/internal/port"
	"notification/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(reader *kafka.Reader) *Consumer {
	return &Consumer{reader: reader}
}

func (c *Consumer) Consume(ctx context.Context, handle func(event domain.TaskEvent) error) error {
	logger.Log(logger.LevelInfo, "Kafka consumer started", nil)

	for {
		select {
		case <-ctx.Done():
			logger.Log(logger.LevelInfo, "Kafka consumer shutting down", nil)
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				logger.Log(logger.LevelError, "Failed to read message from Kafka", err.Error())
				time.Sleep(time.Second)
				continue
			}

			logger.Log(logger.LevelInfo, "Kafka message received", map[string]any{
				"topic": msg.Topic, "partition": msg.Partition, "offset": msg.Offset,
			})

			var event domain.TaskEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logger.Log(logger.LevelError, "Failed to parse Kafka message", map[string]any{
					"error": err.Error(), "payload": string(msg.Value),
				})
				continue
			}

			start := time.Now()
			if err := handle(event); err != nil {
				logger.Log(logger.LevelError, "Event handler error", map[string]any{
					"event": event.Event, "task_id": event.TaskID, "error": err.Error(),
				})
			} else {
				logger.Log(logger.LevelInfo, "Event handled successfully", map[string]any{
					"event": event.Event, "task_id": event.TaskID, "duration": time.Since(start).String(),
				})
			}
		}
	}
}

var _ port.EventConsumer = (*Consumer)(nil)
