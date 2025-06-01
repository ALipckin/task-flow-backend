package controllers_test

import (
	"AuthService/controllers"
	"AuthService/models"
	"bytes"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestSignUp_TableDriven(t *testing.T) {
	db := setupTestDB(t)
	ctrl := &controllers.AuthController{DB: db}

	tests := []struct {
		name       string
		payload    controllers.RequestBody
		wantStatus int
	}{
		{
			name: "Valid signup",
			payload: controllers.RequestBody{
				Email:    "signup1@example.com",
				Name:     "Alice",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Duplicate email",
			payload: controllers.RequestBody{
				Email:    "signup1@example.com",
				Name:     "AliceClone",
				Password: "anotherPass",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing password",
			payload: controllers.RequestBody{
				Email: "signup2@example.com",
				Name:  "Bob",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing email",
			payload: controllers.RequestBody{
				Name:     "Charlie",
				Password: "abc123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Missing name",
			payload: controllers.RequestBody{
				Email:    "signup3@example.com",
				Password: "abc123",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
			w := httptest.NewRecorder()

			ctrl.SignUp(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestLogin_TableDriven(t *testing.T) {
	_ = os.Setenv("JWT_SECRET", "testsecret")

	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
	}{
		{"Valid login", "john@example.com", "secure123", http.StatusOK},
		{"Wrong password", "john@example.com", "wrongpass", http.StatusBadRequest},
		{"Unknown user", "notfound@example.com", "secure123", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)

			if tt.email == "john@example.com" {
				hashed, _ := bcrypt.GenerateFromPassword([]byte("secure123"), 10)
				db.Create(&models.User{
					Email:    "john@example.com",
					Password: string(hashed),
					Name:     "John_" + tt.name, // уникальный name
				})
			}

			ctrl := &controllers.AuthController{DB: db}

			payload := controllers.RequestBody{
				Email:    tt.email,
				Password: tt.password,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			ctrl.Login(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("%s: expected %d, got %d", tt.name, tt.wantStatus, w.Code)
			}
		})
	}
}
