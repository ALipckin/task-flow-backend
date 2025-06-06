package controllers_test

import (
	"AuthService/controllers"
	"AuthService/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&models.User{Email: "alice@example.com", Name: "Alice"})

	ctrl := controllers.UserController{DB: db}

	tests := []struct {
		name       string
		userID     string
		wantStatus int
		wantName   string
	}{
		{"Valid ID", "1", http.StatusOK, "Alice"},
		{"Invalid ID", "999", http.StatusNotFound, ""},
		{"Bad ID format", "abc", http.StatusInternalServerError, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/user?id="+tt.userID, nil)
			w := httptest.NewRecorder()

			ctrl.GetUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var resp controllers.Response
				err := json.NewDecoder(w.Body).Decode(&resp)
				if err != nil {
					t.Fatal("Failed to decode JSON:", err)
				}
				if resp.Name != tt.wantName {
					t.Errorf("expected name %s, got %s", tt.wantName, resp.Name)
				}
			}
		})
	}
}

func TestGetUsers(t *testing.T) {
	db := setupTestDB(t)

	users := []models.User{
		{Email: "u1@example.com", Name: "User1", Group: "user"},
		{Email: "u2@example.com", Name: "User2", Group: "user"},
		{Email: "admin@example.com", Name: "Admin", Group: "admin"},
	}
	for _, u := range users {
		db.Create(&u)
	}

	ctrl := controllers.UserController{DB: db}

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantCount  int
	}{
		{"All non-admin users", "", http.StatusOK, 2},
		{"Search by name", "?search=User1", http.StatusOK, 1},
		{"Filter by ids", "?ids=1,2", http.StatusOK, 2},
		{"Invalid limit", "?limit=bad", http.StatusBadRequest, 0}, // This should return 400
		{"Pagination page 1, limit 1", "?page=1&limit=1", http.StatusOK, 1},
		{"Empty ids", "?ids=,", http.StatusOK, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			w := httptest.NewRecorder()

			ctrl.GetUsers(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
				t.Logf("response body: %s", w.Body.String()) // Add this to debug
			}

			if tt.wantStatus == http.StatusOK {
				var users []controllers.Response
				err := json.NewDecoder(w.Body).Decode(&users)
				if err != nil {
					t.Fatal("Failed to decode response:", err)
				}
				if len(users) != tt.wantCount {
					t.Errorf("expected %d users, got %d", tt.wantCount, len(users))
				}
			}
		})
	}
}
