package usecase

import (
	"context"
	"encoding/json"
	"notification/internal/domain"
	"notification/internal/port"
)

type SendNotification struct {
	userProvider port.UserProvider
	sender       port.Sender
	publisher    port.MessagePublisher
	repository   port.Repository
}

func NewSendNotification(
	userProvider port.UserProvider,
	sender port.Sender,
	publisher port.MessagePublisher,
	repository port.Repository,
) *SendNotification {
	return &SendNotification{
		userProvider: userProvider,
		sender:       sender,
		publisher:    publisher,
		repository:   repository,
	}
}

func (uc *SendNotification) Execute(ctx context.Context, event domain.TaskEvent) error {
	recipients := make(map[int]struct{})
	for _, id := range event.ObserversIDs {
		recipients[id] = struct{}{}
	}
	recipients[event.PerformerID] = struct{}{}
	recipients[event.CreatorID] = struct{}{}

	userIDs := make([]int, 0, len(recipients))
	for id := range recipients {
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		return nil
	}

	users, err := uc.userProvider.GetByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	payload := domain.NotificationPayload{
		Event:       event.Event,
		Title:       event.Title,
		Description: event.Description,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	payloadStr := string(payloadBytes)

	for _, user := range users {
		msg := domain.OutboundMessage{
			UserID:  user.ID,
			Email:   user.Email,
			Message: payloadStr,
		}
		_ = uc.publisher.Publish(ctx, msg)
		_ = uc.sender.Send(user.Email, event.Event, event.Description)
		if uc.repository != nil {
			_ = uc.repository.Save(ctx, &port.NotificationRecord{
				UserID:  user.ID,
				Email:   user.Email,
				Event:   event.Event,
				Payload: payloadStr,
				Channel: "email",
			})
		}
	}

	return nil
}
