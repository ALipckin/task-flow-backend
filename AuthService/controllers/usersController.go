package controllers

import (
	"AuthService/initializers"
	"AuthService/models"
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

type Response struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func GetUser(w http.ResponseWriter, r *http.Request) {
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
		Name:  user.Name,
	}

	json.NewEncoder(w).Encode(userResponse)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	ids := queryParams.Get("ids")
	idList := strings.Split(ids, ",")
	var users []models.User
	if err := initializers.DB.Where("id IN ?", idList).Find(&users).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var userResponses []Response
	for _, user := range users {
		userResponses = append(userResponses, Response{
			Id:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		})
	}

	json.NewEncoder(w).Encode(userResponses)
}
