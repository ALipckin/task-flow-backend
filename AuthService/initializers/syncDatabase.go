package initializers

import (
	"AuthService/models"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func SyncDatabase() {
	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		return
	}

	var count int64
	DB.Model(&models.User{}).Count(&count)
	hash, _ := bcrypt.GenerateFromPassword([]byte("VqjHgT[b6F"), 10)
	if count == 0 {
		user := models.User{
			Email:    "admin@admin.admin",
			Group:    "admin",
			Password: string(hash),
		}

		if err := DB.Create(&user).Error; err != nil {
			log.Fatal("Ошибка при создании пользователя:", err)
		} else {
			fmt.Println("Администратор создан успешно.")
		}
	} else {
		fmt.Println("Администратор уже существует.")
	}
}
