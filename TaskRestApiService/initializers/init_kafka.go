package initializers

import (
	"TaskRestApiService/consumers"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

var KafkaProducer sarama.SyncProducer

func InitProducer() {
	config := newProducerConfig()
	brokers := getKafkaBrokers()

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	KafkaProducer = producer
	log.Println("Kafka producer initialized successfully")
}

func InitConsumer() {
	config := newConsumerConfig()
	brokers := getKafkaBrokers()

	consumer, err := sarama.NewConsumer(brokers, config)
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
