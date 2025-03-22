package models

type TaskEvent struct {
	Event        string `json:"event"`
	TaskID       int    `json:"task_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PerformerID  int    `json:"performer_id"`
	CreatorID    int    `json:"creator_id"`
	ObserversIDs []int  `json:"observers_ids"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
