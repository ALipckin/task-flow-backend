package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBroker      string
	KafkaTopic       string
	KafkaGroupID     string
	KafkaNotifyTopic string
	SMTPHost         string
	SMTPPort         string
	SenderEmail      string
	SenderPassword   string
	AuthServiceURL   string
	AuthToken        string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("loading .env: %v", err)
	}

	cfg := &Config{
		KafkaBroker:      os.Getenv("KAFKA_HOST"),
		KafkaTopic:       os.Getenv("KAFKA_TASK_TOPIC"),
		KafkaGroupID:     os.Getenv("KAFKA_GROUP_ID"),
		KafkaNotifyTopic: os.Getenv("KAFKA_NOTIFY_TOPIC"),
		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPPort:         os.Getenv("SMTP_PORT"),
		SenderEmail:      os.Getenv("SENDER_EMAIL"),
		SenderPassword:   os.Getenv("SENDER_PASSWORD"),
		AuthServiceURL:   os.Getenv("AUTH_SERVICE_URL"),
		AuthToken:        os.Getenv("AUTH_SERVICE_TOKEN"),
	}

	if cfg.KafkaBroker == "" {
		return nil, ErrKafkaHostMissing
	}

	return cfg, nil
}
