package kafka

import (
	"encoding/json"
	"notification/internal/domain"
)

type taskEventMessage struct {
	Event        string `json:"event"`
	TaskID       int    `json:"task_id"`
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PerformerID  int    `json:"performer_id"`
	CreatorID    int    `json:"creator_id"`
	ObserversIDs []int  `json:"observers_ids"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func decodeTaskEvent(payload []byte) (domain.TaskEvent, error) {
	var msg taskEventMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return domain.TaskEvent{}, err
	}

	taskID := msg.TaskID
	if taskID == 0 {
		taskID = msg.ID
	}

	return domain.TaskEvent{
		Event:        msg.Event,
		TaskID:       taskID,
		Title:        msg.Title,
		Description:  msg.Description,
		PerformerID:  msg.PerformerID,
		CreatorID:    msg.CreatorID,
		ObserversIDs: msg.ObserversIDs,
		Status:       msg.Status,
		CreatedAt:    msg.CreatedAt,
		UpdatedAt:    msg.UpdatedAt,
	}, nil
}
