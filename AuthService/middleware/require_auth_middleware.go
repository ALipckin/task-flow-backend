package middleware

import (
	"AuthService/initializers"
	"AuthService/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := parseAndValidateToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var user models.User
		initializers.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			fmt.Println("USER NOT FOUND")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(models.WithUser(r.Context(), &user))

		next.ServeHTTP(w, r)
	})
}

func userBelongsToGroup(user models.User, groupName string) bool {
	if user.Group == groupName {
		return true
	}
	return false
}

func parseAndValidateToken(r *http.Request) (*jwt.Token, jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil, fmt.Errorf("Unauthorized: No token provided")
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, nil, fmt.Errorf("Unauthorized: Invalid token format")
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("Invalid token claims")
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		fmt.Println("Expired token")
		return nil, nil, fmt.Errorf("Token expired")
	}

	return token, claims, nil
}
