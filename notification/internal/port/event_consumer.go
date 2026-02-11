package port

import (
	"context"
	"notification/internal/domain"
)

type EventConsumer interface {
	Consume(ctx context.Context, handle func(event domain.TaskEvent) error) error
}
