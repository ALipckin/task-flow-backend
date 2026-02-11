package main

import (
	"auth/commands"
	"auth/controllers"
	"auth/initializers"
	"auth/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	initializers.LoadEnvVariables()
	DB = initializers.SetUpDB()
	initializers.SyncDatabase(DB)
}

func main() {
	token, err := commands.GenerateAdminToken()
	if err != nil {
		return
	}
	fmt.Println("Auth service token:", token)

	mux := http.NewServeMux()

	authController := controllers.AuthController{
		DB: DB,
	}
	userController := controllers.UserController{
		DB: DB,
	}

	authMiddleware := middleware.RequireAuth(DB)

	mux.HandleFunc("/login", authController.Login)
	mux.HandleFunc("/register", authController.SignUp)

	mux.Handle("/validate", authMiddleware(http.HandlerFunc(authController.Validate)))

	mux.Handle("/user", authMiddleware(http.HandlerFunc(userController.GetUser)))
	mux.Handle("/users", authMiddleware(http.HandlerFunc(userController.GetUsers)))

	server := &http.Server{
		Addr:         ":8081",
		Handler:      middleware.LoggerMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		fmt.Println("Server started on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating graceful shutdown...", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	sqlDB, err := DB.DB()
	if err == nil {
		log.Println("Closing database connection...")
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	log.Println("auth shutdown complete")
}
