package controllers

import (
	"AuthService/initializers"
	"AuthService/logger"
	"AuthService/models"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strconv"
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
	logger.Log(logger.LevelInfo, "Sign-up request received", map[string]any{
		"method": r.Method,
		"ip":     r.RemoteAddr,
	})

	if r.Method != http.MethodPost {
		logger.Log(logger.LevelWarn, "Invalid method for SignUp", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Log(logger.LevelWarn, "Failed to parse JSON body", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	if body.Email == "" || body.Name == "" || body.Password == "" {
		logger.Log(logger.LevelWarn, "Missing required fields in sign-up", body)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
		return
	}

	if len(body.Name) < 3 || len(body.Password) < 6 {
		logger.Log(logger.LevelWarn, "Validation failed for sign-up", map[string]any{
			"name_length":     len(body.Name),
			"password_length": len(body.Password),
		})
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid name or password length"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to hash password", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	user := models.User{Email: body.Email, Password: string(hash), Name: body.Name}
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		logger.Log(logger.LevelError, "Database error during user creation", result.Error.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to create user"})
		return
	}

	logger.Log(logger.LevelInfo, "User created successfully", map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "User created successfully"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	logger.Log(logger.LevelInfo, "Login request received", map[string]any{
		"method": r.Method,
		"ip":     r.RemoteAddr,
	})

	if r.Method != http.MethodPost {
		logger.Log(logger.LevelWarn, "Invalid method for Login", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Log(logger.LevelWarn, "Failed to parse JSON body in login", err.Error())
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		logger.Log(logger.LevelWarn, "Login failed: user not found", body.Email)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email or password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		logger.Log(logger.LevelWarn, "Login failed: incorrect password", body.Email)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   user.Email,
		"user_id": user.ID,
		"name":    user.Name,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		logger.Log(logger.LevelError, "Failed to sign JWT token", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create token"})
		return
	}

	logger.Log(logger.LevelInfo, "Login successful", map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Login successful",
		"token":   tokenString,
		"email":   user.Email,
		"name":    user.Name,
		"id":      strconv.Itoa(int(user.ID)),
	})
}

func Validate(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logger.Log(logger.LevelWarn, "Missing Authorization header", nil)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "No token provided"})
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		logger.Log(logger.LevelWarn, "Malformed token", authHeader)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token format"})
		return
	}

	tokenString := authHeader[7:]
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		logger.Log(logger.LevelWarn, "Token validation failed", err.Error())
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		return
	}

	logger.Log(logger.LevelInfo, "Token validated successfully", claims)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Token is valid",
		"id":      claims["user_id"],
		"email":   claims["email"],
		"name":    claims["name"],
	})
}
