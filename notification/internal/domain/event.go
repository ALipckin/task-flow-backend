package domain

type TaskEvent struct {
	Event        string
	TaskID       int
	Title        string
	Description  string
	PerformerID  int
	CreatorID    int
	ObserversIDs []int
	Status       string
	CreatedAt    string
	UpdatedAt    string
}

const (
	EventTaskCreated = "TaskCreated"
	EventTaskUpdated = "TaskUpdated"
	EventTaskDeleted = "TaskDeleted"
)
