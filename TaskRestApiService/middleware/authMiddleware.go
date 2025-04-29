package middleware

import (
	"TaskRestApiService/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireAuth(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	user, err := services.ValidateToken(authToken)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}
	c.Set("user_id", user.ID)
	c.Next()
}
