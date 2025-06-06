package controllers

import (
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

type UserController struct {
	DB *gorm.DB
}

func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	id := queryParams.Get("id")
	var user models.User
	if err := uc.DB.First(&user, id).Error; err != nil {
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

func (uc UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	ids := queryParams.Get("ids")
	search := queryParams.Get("search")
	pageStr := queryParams.Get("page")
	limitStr := queryParams.Get("limit")

	var users []models.User
	db := uc.DB.Where("\"group\" != ?", "admin")

	if ids != "" {
		idList := strings.Split(ids, ",")
		var validIDs []string
		for _, id := range idList {
			if id != "" {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) > 0 {
			db = db.Where("id IN ?", validIDs)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Response{})
			return
		}
	}

	if search != "" {
		db = db.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if pageStr != "" || limitStr != "" {
		page := 1
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err != nil || p < 1 {
				http.Error(w, "Invalid page parameter", http.StatusBadRequest)
				return
			}
			page = p
		}
		limit := 10
		if limitStr != "" {
			l, err := strconv.Atoi(limitStr)
			if err != nil || l < 1 {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			limit = l
		}
		offset := (page - 1) * limit
		db = db.Offset(offset).Limit(limit)
	}

	if err := db.Find(&users).Error; err != nil {
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
