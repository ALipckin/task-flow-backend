package services

import (
	"NotifyService/initializers"
	"NotifyService/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
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
	log.Printf("📩 Sending notification to users: %v\n", recipients)

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
		log.Fatalf("Error marshalling notification payload: %v", err)
	}

	log.Printf("getting data for usersIds: ", userIDs)
	usersData, err := GetUsersData(userIDs)
	if err != nil {
		log.Fatalf("Error getting users data", err)
	}
	for _, user := range usersData {

		kafkaMessage := KafkaMessage{
			UserID:  user.ID,
			Email:   user.Email,
			Message: string(payloadJSON),
		}
		kafkaMessageJSON, err := json.Marshal(kafkaMessage)
		if err != nil {
			log.Fatalf("Error marshaling Kafka message: %v", err)
		}

		kafkaMessageToSend := kafka.Message{
			Key:   []byte(fmt.Sprintf("user_%d", user.ID)),
			Value: kafkaMessageJSON,
		}

		err = writer.WriteMessages(context.Background(), kafkaMessageToSend)
		if err != nil {
			log.Printf("🚨 Failed to send Kafka message for User %d: %v\n", user.ID, err)
		} else {
			log.Printf("✅ Kafka message sent for User %d\n", user.ID)
		}

		err = SendEmail(user.Email, event.Event, event.Description)
		if err != nil {
			log.Printf("🚨 Failed to send email to %s: %v\n", user.Email, err)
		} else {
			log.Printf("✅ Email sent to %s\n", user.Email)
		}
		log.Printf("📢 Total unique users notified: %d\n", len(uniqueUserIDs))
	}
}
