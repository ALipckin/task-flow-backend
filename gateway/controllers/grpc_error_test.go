package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapGRPCError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "invalid argument",
			err:        status.Error(codes.InvalidArgument, "observer cannot be the task performer"),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "observer cannot be the task performer",
		},
		{
			name:       "not found",
			err:        status.Error(codes.NotFound, "task 1 not found"),
			wantStatus: http.StatusNotFound,
			wantMsg:    "task 1 not found",
		},
		{
			name:       "unavailable",
			err:        status.Error(codes.Unavailable, "connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantMsg:    "connection refused",
		},
		{
			name:       "plain error",
			err:        errors.New("connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotMsg := mapGRPCError(tt.err)
			if gotStatus != tt.wantStatus {
				t.Errorf("status = %d, want %d", gotStatus, tt.wantStatus)
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("message = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}

func TestRespondGRPCError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	respondGRPCError(c, status.Error(codes.InvalidArgument, "observer cannot be the task performer"))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["message"] != "observer cannot be the task performer" {
		t.Fatalf("message = %q, want observer cannot be the task performer", body["message"])
	}
}
