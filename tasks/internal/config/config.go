package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config holds application configuration values.
type Config struct {
	GRPCPort       string
	DBShardURLs    []string
	RedisURL       string
	KafkaBrokers   []string
	VNodesPerShard int
}

var (
	appConfig *Config
	loadOnce  sync.Once
)

// AppConfig returns the singleton config loaded from environment variables.
func AppConfig() *Config {
	loadOnce.Do(func() {
		appConfig = LoadFromEnv()
	})

	return appConfig
}

// LoadFromEnv builds config using environment variables with sane defaults.
func LoadFromEnv() *Config {
	grpcPort := getEnv("GRPC_PORT", "50051")
	redisURL := getEnv("REDIS_URL", "")
	shardURLs := splitEnvList(getEnv("DB_SHARD_URLS", ""))
	kafkaBrokers := splitEnvList(getEnv("KAFKA_BROKERS", "kafka:9092"))
	vnodes := getEnvInt("VNODES_PER_SHARD", 256)

	return &Config{
		GRPCPort:       grpcPort,
		DBShardURLs:    shardURLs,
		RedisURL:       redisURL,
		KafkaBrokers:   kafkaBrokers,
		VNodesPerShard: vnodes,
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func splitEnvList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
