package commands

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

func GenerateAdminToken() (string, error) {
	authTokenKey := "AUTH_SERVICE_TOKEN"
	tokenString := os.Getenv(authTokenKey)
	emailString := os.Getenv("ADMIN_EMAIL")

	if tokenString == "" {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email":   emailString,
			"user_id": 1,
			"exp":     time.Now().Add(time.Hour * 24 * 999).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		err := WriteToEnv(authTokenKey, tokenString)
		if err != nil {
			return "", err
		}
		return tokenString, nil
	} else {
		return tokenString, nil
	}
}
func WriteToEnv(key, value string) error {
	// Открываем файл .env (или создаём, если его нет)
	file, err := os.OpenFile(".env", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Записываем переменную окружения в .env
	_, err = file.WriteString(fmt.Sprintf("%s=%s\n", key, "\""+value+"\""))
	return err
}
