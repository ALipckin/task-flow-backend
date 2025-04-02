package controllers

import (
	"AuthService/initializers"
	"AuthService/models"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

type RequestBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	// Валидация входных данных
	if body.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email is required"})
		return
	}

	if body.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name is required"})
		return
	}

	if len(body.Name) < 3 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Name must be at least 3 characters long"})
		return
	}

	if body.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is required"})
		return
	}

	if len(body.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password must be at least 6 characters long"})
		return
	}

	// Хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	// Создание пользователя
	user := models.User{Email: body.Email, Password: string(hash), Name: body.Name}
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to create user"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "User created successfully"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email or password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   user.Email,
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create token"})
		return
	}
	fmt.Println("creating cooke")
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		Path:     "/",
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Login successful"})
}

// Validate возвращает информацию о пользователе (заглушка)
func Validate(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No token provided"})
		return
	}

	tokenString := cookie.Value
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Token is valid", "id": claims["user_id"], "email": claims["email"]})
}

func Logout(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logout successful"})
}
