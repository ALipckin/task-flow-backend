package services

import (
	"NotifyService/initializers"
	"NotifyService/logger"
	"NotifyService/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
)

type KafkaMessage struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type NotificationPayload struct {
	Event       string `json:"event"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NotifyUsers(event models.TaskEvent) {
	recipients := append(event.ObserversIDs, event.PerformerID, event.CreatorID)

	logger.Log(logger.LevelInfo, "Sending notification to users", map[string]any{
		"recipients": recipients,
	})

	uniqueUserIDs := make(map[int]struct{})
	for _, userID := range recipients {
		uniqueUserIDs[userID] = struct{}{}
	}

	userIDs := make([]int, 0, len(uniqueUserIDs))
	for id := range uniqueUserIDs {
		userIDs = append(userIDs, id)
	}

	writer := initializers.Writer

	notificationPayload := NotificationPayload{
		Event:       event.Event,
		Title:       event.Title,
		Description: event.Description,
	}
	payloadJSON, err := json.Marshal(notificationPayload)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to marshal notification payload", map[string]any{
			"error": err.Error(),
		})
		return
	}

	logger.Log(logger.LevelInfo, "Getting data for user IDs", map[string]any{
		"user_ids": userIDs,
	})

	usersData, err := GetUsersData(userIDs)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to get users data", map[string]any{
			"error": err.Error(),
		})
		return
	}

	for _, user := range usersData {
		kafkaMessage := KafkaMessage{
			UserID:  user.ID,
			Email:   user.Email,
			Message: string(payloadJSON),
		}
		kafkaMessageJSON, err := json.Marshal(kafkaMessage)
		if err != nil {
			logger.Log(logger.LevelError, "Failed to marshal Kafka message", map[string]any{
				"error":   err.Error(),
				"user_id": user.ID,
			})
			continue
		}

		kafkaMessageToSend := kafka.Message{
			Key:   []byte(fmt.Sprintf("user_%d", user.ID)),
			Value: kafkaMessageJSON,
		}

		err = writer.WriteMessages(context.Background(), kafkaMessageToSend)
		if err != nil {
			logger.Log(logger.LevelError, "Failed to send Kafka message", map[string]any{
				"user_id": user.ID,
				"error":   err.Error(),
				"message": kafkaMessageToSend,
			})
		} else {
			logger.Log(logger.LevelInfo, "Kafka message sent", map[string]any{
				"user_id": user.ID,
				"message": kafkaMessageToSend,
			})
		}

		err = SendEmail(user.Email, event.Event, event.Description)
		if err != nil {
			logger.Log(logger.LevelError, "Failed to send email", map[string]any{
				"email": user.Email,
				"error": err.Error(),
			})
		} else {
			logger.Log(logger.LevelInfo, "Email sent", map[string]any{
				"email": user.Email,
			})
		}
	}

	logger.Log(logger.LevelInfo, "Total unique users notified", map[string]any{
		"count": len(uniqueUserIDs),
	})
}
