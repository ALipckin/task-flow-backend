package initializers

import (
	"github.com/gin-contrib/cors"
	"log"
	"os"
	"strings"
	"time"
)

func LoadCorsConfig() cors.Config {
	allowOrigins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	allowMethods := strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ",")
	allowHeaders := strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ",")
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"
	maxAge := os.Getenv("CORS_MAX_AGE")

	maxAgeDuration, err := time.ParseDuration(maxAge + "s")
	if err != nil {
		log.Fatalf("Invalid CORS_MAX_AGE value: %v", err)
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAgeDuration,
	}
	return corsConfig
}
