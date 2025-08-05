package controllers

import (
	"NotifyService/logger"
	"NotifyService/models"
	"NotifyService/services"
	"log"
)

func HandleEvent(event models.TaskEvent) {
	logger.Log(logger.LevelInfo, "Got Event", map[string]any{
		"event":   event.Event,
		"task_id": event.TaskID,
		"title":   event.Title,
	})
	switch event.Event {
	case "TaskCreated":
		services.NotifyUsers(event)
	case "TaskUpdated":
		services.NotifyUsers(event)
	case "TaskDeleted":
		services.NotifyUsers(event)

	default:
		log.Printf("⚠️ Unknown event type: %s\n", event.Event)
	}
}
