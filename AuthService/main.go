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
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/register", controllers.SignUp)
	http.Handle("/validate", middleware.RequireAuth(http.HandlerFunc(controllers.Validate)))
	http.Handle("/logout", middleware.RequireAuth(http.HandlerFunc(controllers.Logout)))

	//middleware.RequireAuthWithGroup("admin", http.HandlerFunc(controllers.GetUser)))
	http.Handle("/user", middleware.RequireAuth(http.HandlerFunc(controllers.GetUser)))
	http.Handle("/users", middleware.RequireAuth(http.HandlerFunc(controllers.GetUsers)))

	fmt.Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}
