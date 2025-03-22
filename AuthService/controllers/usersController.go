package controllers

import (
	"AuthService/initializers"
	"AuthService/models"
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
)

type Response struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	id := queryParams.Get("id")
	var user models.User
	if err := initializers.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userResponse := Response{
		Id:    user.ID,
		Email: user.Email,
	}

	json.NewEncoder(w).Encode(userResponse)
}
