package services

import (
	"NotifyService/initializers"
	"NotifyService/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func getUserData(userId int) map[string]interface{} {
	host := os.Getenv("AUTH_SERVICE_HOST")
	authToken := os.Getenv("AUTH_SERVICE_TOKEN")
	url := host + "/user?id=" + strconv.Itoa(userId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("creating request error:", err)
		return nil
	}

	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error read answer:", err)
		return nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error pars JSON:", err)
		return nil
	}

	return result
}

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

	for userID := range uniqueUserIDs {
		userData := getUserData(userID)

		email, ok := userData["email"].(string)
		if !ok || email == "" {
			log.Printf("⚠️ User %d has no valid email\n", userID)
			continue
		}

		kafkaMessage := KafkaMessage{
			UserID:  userID,
			Email:   email,
			Message: string(payloadJSON),
		}
		kafkaMessageJSON, err := json.Marshal(kafkaMessage)
		if err != nil {
			log.Fatalf("Error marshaling Kafka message: %v", err)
		}

		kafkaMessageToSend := kafka.Message{
			Key:   []byte(fmt.Sprintf("user_%d", userID)),
			Value: kafkaMessageJSON,
		}

		err = writer.WriteMessages(context.Background(), kafkaMessageToSend)
		if err != nil {
			log.Printf("🚨 Failed to send Kafka message for User %d: %v\n", userID, err)
		} else {
			log.Printf("✅ Kafka message sent for User %d\n", userID)
		}

		err = SendEmail(email, event.Event, event.Description)
		if err != nil {
			log.Printf("🚨 Failed to send email to %s: %v\n", email, err)
		} else {
			log.Printf("✅ Email sent to %s\n", email)
		}
	}

	log.Printf("📢 Total unique users notified: %d\n", len(uniqueUserIDs))
}
