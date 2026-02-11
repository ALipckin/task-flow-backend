package config

import "errors"

var ErrKafkaHostMissing = errors.New("KAFKA_HOST env is required")
