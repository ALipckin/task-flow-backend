package consumers

import (
	"NotifyService/controllers"
	"NotifyService/initializers"
	"NotifyService/models"
	"context"
	"encoding/json"
	"log"
)

func StartKafkaConsumer() {
	defer initializers.Reader.Close()

	log.Println("Kafka Consumer started, awaiting messages...")

	for {
		msg, err := initializers.Reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Error reading Kafka message: %v", err)
		}

		log.Printf("📥 Received Kafka message: %s\n", string(msg.Value))

		// Parse the JSON message
		var event models.TaskEvent
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			log.Printf("❌ Error parsing message: %v\n", err)
			continue
		}

		// Handle the event
		controllers.HandleEvent(event)
	}
}
