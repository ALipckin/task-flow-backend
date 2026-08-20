package controllers

import (
	"errors"
	"testing"
)

func TestUserCreateError(t *testing.T) {
	tests := []struct {
		err  string
		want string
	}{
		{
			err:  `ERROR: duplicate key value violates unique constraint "uni_users_email" (SQLSTATE 23505)`,
			want: "Email already exists",
		},
		{
			err:  `ERROR: duplicate key value violates unique constraint "uni_users_name" (SQLSTATE 23505)`,
			want: "Name already exists",
		},
		{
			err:  "UNIQUE constraint failed: users.email",
			want: "Email already exists",
		},
		{
			err:  "UNIQUE constraint failed: users.name",
			want: "Name already exists",
		},
		{
			err:  "connection refused",
			want: "Failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.err, func(t *testing.T) {
			got := userCreateError(errors.New(tt.err))
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
