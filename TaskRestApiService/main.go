package main

import (
	"TaskRestApiService/consumers"
	"TaskRestApiService/controllers"
	_ "TaskRestApiService/docs"
	"TaskRestApiService/initializers"
	"TaskRestApiService/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
	"os"
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
	defer initializers.KafkaProducer.Close()
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

	authHost := os.Getenv("AUTH_SERVICE_URL")
	authController := controllers.NewAuthController(authHost)

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
	r.Run(":" + port)
}
