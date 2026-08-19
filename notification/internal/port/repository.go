package port

import (
	"context"
	"notification/internal/domain"
)

type Repository interface {
	Save(ctx context.Context, n *domain.Notification) error
}
