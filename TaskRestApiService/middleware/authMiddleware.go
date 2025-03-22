package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"time"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func RequireAuth(c *gin.Context) {
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")

	// Проверяем наличие cookie с токеном
	_, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token"})
		return
	}

	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", authServiceURL+"/validate", nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Переносим все cookies из запроса в новый запрос к auth-сервису
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	// Отправляем запрос на аутентификацию
	resp, err := client.Do(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Auth service error"})
		return
	}
	defer resp.Body.Close()

	// Если статус не 200, считаем токен невалидным
	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid token"})
		return
	}
	log.Printf("resp: ", resp.Body)
	// Десериализуем тело ответа в объект пользователя
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid response from auth service"})
		return
	}

	// Сохраняем информацию о пользователе в контексте запроса
	c.Set("user_id", user.ID)

	// Продолжаем выполнение следующего обработчика
	c.Next()
}
