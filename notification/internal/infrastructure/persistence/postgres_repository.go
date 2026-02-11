package persistence

import (
	"context"
	"notification/internal/port"
)

type PostgresRepository struct{}

func NewPostgresRepository( /* ds string or *sql.DB */ ) *PostgresRepository {
	return &PostgresRepository{}
}

func (r *PostgresRepository) Save(ctx context.Context, rec *port.NotificationRecord) error {
	return nil
}

var _ port.Repository = (*PostgresRepository)(nil)
