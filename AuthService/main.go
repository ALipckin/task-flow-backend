package main

import (
	"AuthService/commands"
	"AuthService/controllers"
	"AuthService/initializers"
	"AuthService/middleware"
	"fmt"
	"gorm.io/gorm"
	"net/http"
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

	fmt.Println("Server started on :8081")
	http.ListenAndServe(":8081", middleware.LoggerMiddleware(mux))
}
