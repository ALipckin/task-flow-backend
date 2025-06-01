package initializers

import (
	"AuthService/models"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"os"
)

func SyncDatabase(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Ошибка миграции:", err)
	}

	var count int64
	db.Model(&models.User{}).Count(&count)
	passwordString := os.Getenv("ADMIN_PASS")
	hash, _ := bcrypt.GenerateFromPassword([]byte(passwordString), 10)

	if count == 0 {
		user := models.User{
			Email:    "admin@admin.admin",
			Group:    "admin",
			Name:     "admin",
			Password: string(hash),
		}

		if err := db.Create(&user).Error; err != nil {
			log.Fatal("Ошибка при создании пользователя:", err)
		} else {
			fmt.Println("Администратор создан успешно.")
		}
	} else {
		fmt.Println("Администратор уже существует.")
	}
}
