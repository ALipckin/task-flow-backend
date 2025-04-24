package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
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

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid token format"})
		return
	}

	//tokenString := authHeader[7:]

	client := http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", authServiceURL+"/validate", nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Auth service error"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid token"})
		return
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid response from auth service"})
		return
	}

	c.Set("user_id", user.ID)

	c.Next()
}
