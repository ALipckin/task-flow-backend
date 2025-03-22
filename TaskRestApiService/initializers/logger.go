package initializers

import (
	"encoding/json"
	"log"
)

// LogMessage represents the log structure
type LogMessage struct {
	Level   string      `json:"level"`   // Log level: info, error
	Action  string      `json:"action"`  // Action name (e.g., TasksCreate)
	Message string      `json:"message"` // Description of the event
	Details interface{} `json:"details"` // Additional details (ID, errors, etc.)
}

// LogToKafka sends logs to Kafka
func LogToKafka(level, action, message string, details interface{}) {
	logEntry := LogMessage{
		Level:   level,
		Action:  action,
		Message: message,
		Details: details,
	}

	logData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Error serializing log: %v", err)
		return
	}

	err = SendMessage("service-logs", string(logData))
	if err != nil {
		log.Printf("Error sending log to Kafka: %v", err)
	}
}
