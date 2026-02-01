package initializers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	RedisClient = GetClient()

	// Verify connection
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Println(fmt.Sprintf("Redis connection error: %v", err))
		panic(err)
	}
	log.Println("Redis connected")
}

func GetClient() *redis.Client {
	url := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return redis.NewClient(opts)
}
