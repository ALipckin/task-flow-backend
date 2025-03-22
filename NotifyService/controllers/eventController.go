package controllers

import (
	"NotifyService/models"
	"NotifyService/services"
	"log"
)

func HandleEvent(event models.TaskEvent) {
	switch event.Event {
	case "TaskCreated":
		log.Printf("🆕 Task Created: ID: %s, Title: %s\n", event.TaskID, event.Title)
		services.NotifyUsers(event)

	case "TaskUpdated":
		log.Printf("✏️ Task Updated: ID: %s, Title: %s\n", event.TaskID, event.Title)
		services.NotifyUsers(event)

	case "TaskDeleted":
		log.Printf("❌ Task Deleted: ID: %s\n", event.TaskID)
		services.NotifyUsers(event)

	default:
		log.Printf("⚠️ Unknown event type: %s\n", event.Event)
	}
}
