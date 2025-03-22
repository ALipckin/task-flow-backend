package services

import (
	"NotifyService/initializers"
	"NotifyService/models"
	"bytes"
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

func login() (string, error) {
	fmt.Println("login")
	email := os.Getenv("AUTH_SERVICE_ADMIN_LOGIN")
	password := os.Getenv("AUTH_SERVICE_ADMIN_PASSWORD")
	host := os.Getenv("AUTH_SERVICE_HOST")
	url := host + "/login"

	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth error: %d", resp.StatusCode)
	}
	cookie, err := resp.Cookies()[0].Value, err
	if err != nil {
		return "", fmt.Errorf("cookie hollow error: %v", err)
	}
	fmt.Println("return cookie")
	return cookie, nil
}

func getUserData(userId int) map[string]interface{} {
	host := os.Getenv("AUTH_SERVICE_HOST")
	url := host + "/getUserInfo?id=" + strconv.Itoa(userId)

	// Получаем токен с помощью авторизации
	token, err := login()
	if err != nil {
		fmt.Println("auth error:", err)
		return nil
	}

	// Создаем новый GET запрос
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("creating request error:", err)
		return nil
	}

	// Добавляем токен в cookie
	req.AddCookie(&http.Cookie{
		Name:  "Authorization",
		Value: token,
	})

	// Создаем HTTP клиент и отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return nil
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error read answer:", err)
		return nil
	}

	// Декодируем JSON в map
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

func NotifyUsers(event models.TaskEvent) {
	// Collect unique user IDs
	recipients := append(event.ObserversIDs, event.PerformerID, event.CreatorID)
	log.Printf("📩 Sending notification to users: %v\n", recipients)

	uniqueUserIDs := make(map[int]struct{})
	for _, userID := range recipients {
		uniqueUserIDs[userID] = struct{}{}
	}

	log.Printf("📩 Processing %d unique users\n", len(uniqueUserIDs))

	// Initialize Kafka writer (assuming it's configured in initializers)
	writer := initializers.Writer // Kafka Writer instance initialized elsewhere

	// Prepare the message content for both Kafka and email
	messageContent := fmt.Sprintf("Event: %s\nTitle: %s\nDescription: %s", event.Event, event.Title, event.Description)

	// Send notification messages to both Kafka and email
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
			Message: messageContent,
		}
		kafkaMessageJSON, err := json.Marshal(kafkaMessage)
		if err != nil {
			log.Fatalf("Error marshaling Kafka message: %v", err)
		}

		kafkaMessageToSend := kafka.Message{
			Key:   []byte(fmt.Sprintf("user_%d", userID)), // Partition key by user ID
			Value: kafkaMessageJSON,                       // Message content
		}

		err = writer.WriteMessages(context.Background(), kafkaMessageToSend)
		if err != nil {
			log.Printf("🚨 Failed to send Kafka message for User %d: %v\n", userID, err)
		} else {
			log.Printf("✅ Kafka message sent for User %d\n", userID)
		}

		// Send the email notification
		err = SendEmail(email, event.Event, messageContent)
		if err != nil {
			log.Printf("🚨 Failed to send email to %s: %v\n", email, err)
		} else {
			log.Printf("✅ Email sent to %s\n", email)
		}
	}

	log.Printf("📢 Total unique users notified: %d\n", len(uniqueUserIDs))
}
