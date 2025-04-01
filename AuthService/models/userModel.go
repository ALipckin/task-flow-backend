package models

import "gorm.io/gorm"
import (
	"context"
)

type User struct {
	gorm.Model
	ID       int    `gorm:"primaryKey;autoIncrement"`
	Email    string `gorm:"unique;not null"`
	Name     string `gorm:"unique;not null"`
	Group    string `gorm:"default:'user'"`
	Password string `gorm:"not null"`
}

type contextKey string

const userContextKey contextKey = "user"

func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUser(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}
