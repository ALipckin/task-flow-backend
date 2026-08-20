package initializers

import (
	"fmt"
	"gateway/consumers"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

const (
	kafkaConnectAttempts = 60
	kafkaConnectInterval = 2 * time.Second
)

var KafkaProducer sarama.SyncProducer

func InitProducer() {
	config := newProducerConfig()
	brokers := getKafkaBrokers()

	producer, err := connectProducer(brokers, config, kafkaConnectAttempts, kafkaConnectInterval)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	KafkaProducer = producer
	log.Println("Kafka producer initialized successfully")
}

func InitConsumer() {
	config := newConsumerConfig()
	brokers := getKafkaBrokers()

	consumer, err := connectConsumer(brokers, config, kafkaConnectAttempts, kafkaConnectInterval)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}

	consumers.KafkaConsumer = consumer
	log.Println("Kafka consumer initialized successfully")
}

func getKafkaBrokers() []string {
	host := os.Getenv("KAFKA_HOST")
	port := os.Getenv("KAFKA_PORT")
	if host == "" || port == "" {
		log.Fatal("KAFKA_HOST and KAFKA_PORT must be set")
	}
	return []string{fmt.Sprintf("%s:%s", host, port)}
}

func newProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 5 * time.Second
	return config
}

func newConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	return config
}

func connectProducer(
	brokers []string,
	cfg *sarama.Config,
	attempts int,
	interval time.Duration,
) (sarama.SyncProducer, error) {
	return retryKafkaConnect("producer", attempts, interval, func() (sarama.SyncProducer, error) {
		return sarama.NewSyncProducer(brokers, cfg)
	})
}

func connectConsumer(
	brokers []string,
	cfg *sarama.Config,
	attempts int,
	interval time.Duration,
) (sarama.Consumer, error) {
	return retryKafkaConnect("consumer", attempts, interval, func() (sarama.Consumer, error) {
		return sarama.NewConsumer(brokers, cfg)
	})
}

func retryKafkaConnect[T any](kind string, attempts int, interval time.Duration, connect func() (T, error)) (T, error) {
	if attempts < 1 {
		attempts = 1
	}

	var zero T
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		client, err := connect()
		if err == nil {
			return client, nil
		}
		lastErr = err
		if attempt == attempts {
			break
		}
		log.Printf("Kafka %s not ready (attempt %d/%d): %v", kind, attempt, attempts, err)
		if interval > 0 {
			time.Sleep(interval)
		}
	}
	return zero, lastErr
}
