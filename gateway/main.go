package main

import (
	"context"
	"gateway/consumers"
	"gateway/controllers"
	_ "gateway/docs"
	"gateway/initializers"
	"gateway/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	initializers.InitProducer()
	initializers.InitConsumer()
}

// @title Task REST API
// @version 0.1
// @description This is a REST API for managing tasks
// @termsOfService http://swagger.io/terms/

// @BasePath /

// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer {token}"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if initializers.KafkaProducer != nil {
			initializers.KafkaProducer.Close()
		}
	}()

	mode := os.Getenv("GIN_MODE")
	gin.SetMode(mode)

	notifyTopic := os.Getenv("KAFKA_NOTIFY_TOPIC")
	go consumers.ConsumeMessages(notifyTopic)

	r := gin.Default()
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RateLimiterWithConfig(10, 20))

	r.Use(cors.New(initializers.LoadCorsConfig()))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is working",
		})
	})

	grpcClient := initializers.InitTaskStorageService()
	taskController := controllers.NewTaskController(grpcClient)

	tasksGroup := r.Group("/tasks", middleware.RequireAuth)
	{
		tasksGroup.POST("", taskController.TasksCreate)
		tasksGroup.GET("", taskController.TasksIndex)
		tasksGroup.GET("/:id", taskController.TasksShow)
		tasksGroup.PUT("/:id", taskController.TasksUpdate)
		tasksGroup.DELETE("/:id", taskController.TasksDelete)
	}
	r.GET("/tasks/notifications", consumers.HandleWebSocketConnection)

	authUrl := os.Getenv("AUTH_SERVICE_URL")
	authController := controllers.NewAuthController(authUrl)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/register", authController.Register)
		authGroup.GET("/validate", authController.Validate)
		authGroup.GET("/users", authController.Users)
		authGroup.GET("/user", authController.User)
		authGroup.POST("/logout", authController.Logout)
	}

	port := os.Getenv("PORT")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if initializers.KafkaProducer != nil {
		log.Println("Closing Kafka producer...")
		if err := initializers.KafkaProducer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}

	if consumers.KafkaConsumer != nil {
		log.Println("Closing Kafka consumer...")
		if err := consumers.KafkaConsumer.Close(); err != nil {
			log.Printf("Error closing Kafka consumer: %v", err)
		}
	}

	log.Println("gateway shutdown complete")
}
