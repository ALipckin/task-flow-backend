package kafke

import (
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

var KafkaProducer sarama.SyncProducer

// InitProducer initializes the Kafka producer with provided broker list.
func InitProducer(brokers []string) {
	brokers = normalizeBrokers(brokers)
	if len(brokers) == 0 {
		brokers = []string{"kafka:9092"}
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 5 * time.Second

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	KafkaProducer = producer
	log.Println("Kafka producer initialized successfully")
}

// SendMessage sends a message to Kafka.
func SendMessage(topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	if KafkaProducer == nil {
		log.Printf("No Kafka producer available, skipping message: %s", message)
		return nil
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
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	if KafkaProducer == nil {
		log.Printf("No Kafka producer available to send message")
		return nil
	}

	_, _, err := KafkaProducer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return err
	}
	return nil
}

// CloseProducer closes the package producer.
func CloseProducer() error {
	if KafkaProducer == nil {
		return nil
	}
	if err := KafkaProducer.Close(); err != nil {
		return err
	}
	KafkaProducer = nil
	return nil
}

func normalizeBrokers(brokers []string) []string {
	result := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		item := strings.TrimSpace(broker)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
