package domain

import (
	"tasks/internal/infrastructure/persistence"
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID          uint
	Title       string
	Description string
	PerformerId uint
	CreatorId   uint
	Observers   []persistence.Observer
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt
}

// NewTask constructs a domain Task value from parameters.
func NewTask(id uint, title, description string, creatorID, performerID uint) Task {
	return Task{
		ID:          id,
		Title:       title,
		Description: description,
		PerformerId: performerID,
		CreatorId:   creatorID,
		Status:      "new",
	}
}
