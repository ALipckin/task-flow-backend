package port

import "context"

type NotificationRecord struct {
	UserID  int
	Email   string
	Event   string
	Payload string
	SentAt  string
	Channel string
}

type Repository interface {
	Save(ctx context.Context, r *NotificationRecord) error
}
