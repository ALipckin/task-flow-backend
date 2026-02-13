package ports

import (
	"context"
	"tasks/internal/domain"
)

// EventProducer publishes domain events (e.g., task created).
// Use-cases depend on this interface to notify other services.
type EventProducer interface {
	PublishCreated(ctx context.Context, task domain.Task) error
}
