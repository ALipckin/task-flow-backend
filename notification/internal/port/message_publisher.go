package port

import (
	"context"
	"notification/internal/domain"
)

type MessagePublisher interface {
	Publish(ctx context.Context, msg domain.OutboundMessage) error
}
