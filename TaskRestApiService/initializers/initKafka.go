package initializers

import (
	"TaskRestApiService/consumers"
	"github.com/IBM/sarama"
	"log"
	"os"
	"time"
)

var KafkaProducer sarama.SyncProducer

func InitProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true // Wait for confirmation of successful sending
	config.Producer.Timeout = 5 * time.Second
	kafkaHost := os.Getenv("KAFKA_HOST")
	brokers := []string{kafkaHost}

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

func InitConsumer() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true                  // Enable error handling for consumers
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // Start from the newest message

	kafkaHost := os.Getenv("KAFKA_HOST")
	brokers := []string{kafkaHost}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}

	consumers.KafkaConsumer = consumer
	log.Println("Kafka consumer initialized successfully")
}
