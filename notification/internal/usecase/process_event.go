package usecase

import (
	"context"
	"notification/internal/domain"
)

type ProcessEvent struct {
	sendNotification *SendNotification
}

func NewProcessEvent(sendNotification *SendNotification) *ProcessEvent {
	return &ProcessEvent{sendNotification: sendNotification}
}

func (uc *ProcessEvent) Execute(ctx context.Context, event domain.TaskEvent) error {
	switch event.Event {
	case domain.EventTaskCreated, domain.EventTaskUpdated, domain.EventTaskDeleted:
		return uc.sendNotification.Execute(ctx, event)
	default:
		return domain.ErrUnknownEventType
	}
}
