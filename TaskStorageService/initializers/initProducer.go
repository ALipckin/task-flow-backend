package initializers

import (
	"github.com/IBM/sarama"
	"log"
	"time"
)

var KafkaProducer sarama.SyncProducer

// InitProducer initializes the Kafka producer
func InitProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true // Wait for confirmation of successful sending
	config.Producer.Timeout = 5 * time.Second

	// Kafka broker address
	brokers := []string{"kafka:9092"} // Using the service name from docker-compose

	// Create a Kafka producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	KafkaProducer = producer
	log.Println("Kafka producer initialized successfully")
}

// SendMessage sends a message to Kafka
func SendMessage(topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	_, _, err := KafkaProducer.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending message to Kafka: %v", err)
		return err
	}

	log.Printf("Message successfully sent to Kafka: %s", message)
	return nil
}

func SendMessageToKafka(message []byte) error {
	topic := "task_events"
	// Формируем Kafka сообщение
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	// Отправляем сообщение в Kafka
	var err error
	_, _, err = KafkaProducer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return err
	}

	return nil
}
