package controllers

import (
	"AuthService/initializers"
	"AuthService/models"
	"encoding/json"
	"net/http"
	"strconv"
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
	search := queryParams.Get("search")
	page := "0"
	page = queryParams.Get("page")
	limit := "10"
	limit = queryParams.Get("limit")

	var users []models.User
	db := initializers.DB

	db = db.Where("\"group\" != ?", "admin")

	if ids != "" {
		idList := strings.Split(ids, ",")
		// Фильтруем пустые строки из списка id
		var validIDs []string
		for _, id := range idList {
			if id != "" {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) > 0 {
			if err := db.Where("id IN ?", validIDs).Find(&users).Error; err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		} else {
			// Если нет валидных ID, возвращаем пустой список
			users = []models.User{}
		}
	} else {
		if search != "" {
			db = db.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
		}

		if page != "" && limit != "" {
			pageInt, err := strconv.Atoi(page)
			if err != nil {
				http.Error(w, "Invalid page parameter", http.StatusBadRequest)
				return
			}
			limitInt, err := strconv.Atoi(limit)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			offset := (pageInt - 1) * limitInt
			db = db.Offset(offset).Limit(limitInt)
		}

		if err := db.Find(&users).Error; err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
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
