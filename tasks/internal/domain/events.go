package domain

import "time"

type DomainEvent interface {
	Name() string
}

type TaskCreated struct {
	TaskID      uint
	Title       string
	CreatorID   uint
	PerformerID uint
	OccurredAt  time.Time
}

func (e TaskCreated) Name() string { return "TaskCreated" }

type TaskUpdated struct {
	TaskID     uint
	OccurredAt time.Time
}

func (e TaskUpdated) Name() string { return "TaskUpdated" }

type TaskDeleted struct {
	TaskID     uint
	OccurredAt time.Time
}

func (e TaskDeleted) Name() string { return "TaskDeleted" }
