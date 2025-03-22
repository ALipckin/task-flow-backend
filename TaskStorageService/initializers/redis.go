package initializers

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
)

var RedisClient *redis.Client

func ConnectRedis() {
	RedisClient = GetClient()

	// Проверяем соединение
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Println(fmt.Sprintf("Ошибка подключения к Redis: %v", err))
		panic(err)
	}
	log.Println("✅ Подключено к Redis!")
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
