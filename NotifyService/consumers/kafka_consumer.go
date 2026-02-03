package consumers

import (
	"NotifyService/controllers"
	"NotifyService/initializers"
	"NotifyService/logger"
	"NotifyService/models"
	"context"
	"encoding/json"
	"time"
)

func StartKafkaConsumer(ctx context.Context) {
	logger.Log(logger.LevelInfo, "Kafka consumer started", nil)

	for {
		select {
		case <-ctx.Done():
			logger.Log(logger.LevelInfo, "Kafka consumer shutting down", nil)
			return
		default:
			msg, err := initializers.Reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					logger.Log(logger.LevelInfo, "Context cancelled, stopping consumer", nil)
					return
				}
				logger.Log(logger.LevelError, "Failed to read message from Kafka", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}

			logger.Log(logger.LevelInfo, "Kafka message received", map[string]any{
				"topic":     msg.Topic,
				"partition": msg.Partition,
				"offset":    msg.Offset,
				"key":       string(msg.Key),
			})

			var event models.TaskEvent
			err = json.Unmarshal(msg.Value, &event)
			if err != nil {
				logger.Log(logger.LevelError, "Failed to parse Kafka message", map[string]any{
					"error":   err.Error(),
					"payload": string(msg.Value),
				})
				continue
			}

			start := time.Now()
			controllers.HandleEvent(event)
			duration := time.Since(start)

			logger.Log(logger.LevelInfo, "Event handled successfully", map[string]any{
				"event":    event.Event,
				"task_id":  event.TaskID,
				"duration": duration.String(),
			})
		}
	}
}
