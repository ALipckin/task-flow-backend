package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"notification/internal/domain"
	"notification/internal/port"
	"notification/logger"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(writer *kafka.Writer) *Publisher {
	return &Publisher{writer: writer}
}

func (p *Publisher) Publish(ctx context.Context, msg domain.OutboundMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to marshal message", map[string]any{"error": err.Error()})
		return err
	}

	kmsg := kafka.Message{
		Key:   []byte(fmt.Sprintf("user_%d", msg.UserID)),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, kmsg); err != nil {
		logger.Log(logger.LevelError, "Failed to publish message", map[string]any{
			"user_id": msg.UserID, "error": err.Error(),
		})
		return err
	}

	logger.Log(logger.LevelInfo, "Message published", map[string]any{"user_id": msg.UserID})
	return nil
}

var _ port.MessagePublisher = (*Publisher)(nil)
