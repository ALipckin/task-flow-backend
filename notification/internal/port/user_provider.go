package port

import (
	"context"
	"notification/internal/domain"
)

type UserProvider interface {
	GetByIDs(ctx context.Context, ids []int) ([]domain.User, error)
}
