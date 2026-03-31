package persistence

import "time"

type Observer struct {
	ID        uint
	UserId    uint
	TaskId    uint
	Task      Task
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
