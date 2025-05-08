package main

import (
	"TaskRestApiService/consumers"
	"TaskRestApiService/controllers"
	"TaskRestApiService/initializers"
	"TaskRestApiService/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.InitProducer()
	initializers.InitConsumer()
}

func main() {
	defer initializers.KafkaProducer.Close()
	mode := os.Getenv("GIN_MODE")
	gin.SetMode(mode)

	notifyTopic := os.Getenv("KAFKA_NOTIFY_TOPIC")
	go consumers.ConsumeMessages(notifyTopic)

	grpcClient := initializers.InitTaskStorageService()
	taskController := controllers.NewTaskController(grpcClient)

	r := gin.Default()
	r.Use(middleware.LoggerMiddleware())

	allowOrigins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	allowMethods := strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ",")
	allowHeaders := strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ",")
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"
	maxAge := os.Getenv("CORS_MAX_AGE")

	maxAgeDuration, err := time.ParseDuration(maxAge + "s")
	if err != nil {
		log.Fatalf("Invalid CORS_MAX_AGE value: %v", err)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAgeDuration,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is working",
		})
	})

	tasksGroup := r.Group("/tasks", middleware.RequireAuth)
	{
		tasksGroup.POST("", taskController.TasksCreate)
		tasksGroup.GET("", taskController.TasksIndex)
		tasksGroup.GET("/:id", taskController.TasksShow)
		tasksGroup.PUT("/:id", taskController.TasksUpdate)
		tasksGroup.DELETE("/:id", taskController.TasksDelete)
	}
	r.GET("/tasks/notifications", consumers.HandleWebSocketConnection)

	authGroup := r.Group("/auth")
	{
		authHost := os.Getenv("AUTH_SERVICE_URL")

		authGroup.POST("/login", func(c *gin.Context) {
			log.Printf("received request")
			targetURL := c.DefaultQuery("url", authHost+"/login")

			controllers.ProxyRequest(c, targetURL)
		})
		authGroup.POST("/register", func(c *gin.Context) {
			targetURL := c.DefaultQuery("url", authHost+"/register")

			controllers.ProxyRequest(c, targetURL)
		})
		authGroup.GET("/validate", func(c *gin.Context) {
			targetURL := c.DefaultQuery("url", authHost+"/validate")
			controllers.ProxyRequest(c, targetURL)
		})
		authGroup.GET("/users", func(c *gin.Context) {
			targetURL := c.DefaultQuery("url", authHost+"/users")
			controllers.ProxyRequest(c, targetURL)
		})
		authGroup.GET("/user", func(c *gin.Context) {
			targetURL := c.DefaultQuery("url", authHost+"/user")
			controllers.ProxyRequest(c, targetURL)
		})
		authGroup.POST("/logout", func(c *gin.Context) {
			targetURL := c.DefaultQuery("url", authHost+"/logout")
			controllers.ProxyRequest(c, targetURL)
		})
	}

	r.Run()
}
