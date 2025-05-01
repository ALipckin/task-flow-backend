package main

import (
	"AuthService/commands"
	"AuthService/controllers"
	"AuthService/initializers"
	"AuthService/middleware"
	"fmt"
	"net/http"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	token, err := commands.GenerateAdminToken()
	if err != nil {
		return
	}
	fmt.Println("Auth service token: ", token)
	mux := http.NewServeMux()
	mux.HandleFunc("/login", controllers.Login)
	mux.HandleFunc("/register", controllers.SignUp)
	mux.Handle("/validate", middleware.RequireAuth(http.HandlerFunc(controllers.Validate)))

	mux.Handle("/user", middleware.RequireAuth(http.HandlerFunc(controllers.GetUser)))
	mux.Handle("/users", middleware.RequireAuth(http.HandlerFunc(controllers.GetUsers)))

	fmt.Println("Server started on :8081")
	http.ListenAndServe(":8081", middleware.LoggerMiddleware(mux))
}
