package consumers

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var KafkaConsumer sarama.Consumer
var clients = make(map[*websocket.Conn]string) // Список WebSocket клиентов с их user_id

// WebSocket настройки
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type KafkaMessage struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type EventDataMessage struct {
	Event       string `json:"event"`
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      int    `json:"user_id"`
}

func HandleWebSocketConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error connection to WebSocket:", err)
		return
	}
	defer conn.Close()

	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("No user ID found in context")
		conn.Close()
		return
	}

	// Добавляем клиента и его user_id в карту
	clients[conn] = strconv.Itoa(userID.(int))
	log.Println("New WebSocket client connected with user_id:", userID)

	// Ожидаем закрытия соединения
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message", err)
			delete(clients, conn)
			break
		}
	}
}

func BroadcastToClients(message []byte, userID int) {
	log.Println("Broadcasting message to clients")
	for client, clientID := range clients {
		// Отправляем сообщение только клиенту с соответствующим user_id
		if clientID == strconv.Itoa(userID) {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error sending message to WebSocket client:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func ConsumeMessages(topic string) {
	partitionConsumer, err := KafkaConsumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error creating partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	log.Printf("Start consuming message from topic: %s", topic)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Got message: %s", string(msg.Value))

			var raw KafkaMessage
			if err := json.Unmarshal(msg.Value, &raw); err != nil {
				log.Printf("Error unmarshaling Kafka raw message: %v", err)
				continue
			}

			var eventData EventDataMessage
			if err := json.Unmarshal([]byte(raw.Message), &eventData); err != nil {
				log.Printf("Error unmarshaling inner message: %v", err)
				continue
			}

			eventData.UserID = raw.UserID

			jsonMessage, err := json.Marshal(eventData)
			if err != nil {
				log.Printf("Error marshalling final message: %v", err)
				continue
			}

			log.Printf("json: %s", string(jsonMessage))
			BroadcastToClients(jsonMessage, eventData.UserID)

		case err := <-partitionConsumer.Errors():
			log.Printf("Error getting message: %v", err)
		case <-signals:
			log.Println("Closing consumer")
			return
		}
	}
}
