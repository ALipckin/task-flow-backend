package persistence

import (
	"time"

	"gorm.io/gorm"
)

type Observer struct {
	ID        uint `gorm:"primaryKey"`
	UserId    uint `gorm:"not null;index"`
	TaskId    uint `gorm:"not null;index"`
	Task      Task `gorm:"foreignKey:TaskId;references:ID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
