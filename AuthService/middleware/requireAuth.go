package middleware

import (
	"AuthService/initializers"
	"AuthService/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strconv"
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

func RequireAuthWithGroup(groupName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем и проверяем токен из куков
		_, claims, err := parseAndValidateToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Ищем пользователя в базе данных
		var user models.User

		initializers.DB.First(&user, claims["user_id"])

		if user.ID == 0 {
			fmt.Println("USER NOT FOUND")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем, принадлежит ли пользователь нужной группе
		if !userBelongsToGroup(user, groupName) {
			fmt.Println("User " + strconv.Itoa(user.ID) + " does not belong to the required group: userGroup " + user.Group + "!=" + groupName)
			http.Error(w, "Forbidden", http.StatusForbidden)
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
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		return nil, nil, fmt.Errorf("Unauthorized: %v", err)
	}
	tokenString := cookie.Value
	fmt.Println("TokenString:", tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("Invalid token claims")
	}
	// Проверяем срок действия токена
	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		fmt.Println("Expired token")
		return nil, nil, fmt.Errorf("Token expired")
	}
	return token, claims, nil
}
