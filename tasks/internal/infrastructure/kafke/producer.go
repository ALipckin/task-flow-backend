package kafke

import (
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Producer sends messages to Kafka.
type Producer struct {
	sync sarama.SyncProducer
}

// NewProducer creates a Kafka sync producer for the given brokers.
func NewProducer(brokers []string) *Producer {
	brokers = normalizeBrokers(brokers)
	if len(brokers) == 0 {
		brokers = []string{"kafka:9092"}
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Timeout = 5 * time.Second

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	log.Println("Kafka producer initialized successfully")
	return &Producer{sync: producer}
}

func (p *Producer) SendMessage(topic, message string) error {
	if p == nil || p.sync == nil {
		log.Printf("No Kafka producer available, skipping message: %s", message)
		return nil
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	_, _, err := p.sync.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending message to Kafka: %v", err)
		return err
	}
	log.Printf("Message successfully sent to Kafka: %s", message)
	return nil
}

func (p *Producer) SendMessageToKafka(message []byte) error {
	if p == nil || p.sync == nil {
		log.Printf("No Kafka producer available to send message")
		return nil
	}

	msg := &sarama.ProducerMessage{
		Topic: "task_events",
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := p.sync.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		return err
	}
	return nil
}

func (p *Producer) Close() error {
	if p == nil || p.sync == nil {
		return nil
	}
	if err := p.sync.Close(); err != nil {
		return err
	}
	p.sync = nil
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
