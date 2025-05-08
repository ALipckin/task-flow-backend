package logger

import (
	"encoding/json"
	"os"
	"time"
)

type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp  string   `json:"timestamp"`
	Level      LogLevel `json:"level"`
	Message    string   `json:"message"`
	Method     string   `json:"method,omitempty"`
	Path       string   `json:"path,omitempty"`
	RemoteAddr string   `json:"remote_addr,omitempty"`
	UserAgent  string   `json:"user_agent,omitempty"`
	StatusCode int      `json:"status_code,omitempty"`
	DurationMS int64    `json:"duration_ms,omitempty"`
	Context    any      `json:"context,omitempty"`
}

var logEncoder = json.NewEncoder(os.Stdout)

func Log(level LogLevel, message string, ctx any) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Context:   ctx,
	}
	logEncoder.Encode(entry)
}

func LogRequest(method, path, remoteAddr, userAgent string, statusCode int, duration time.Duration) {
	entry := LogEntry{
		Timestamp:  time.Now().Format(time.RFC3339),
		Level:      LevelInfo,
		Message:    "HTTP request",
		Method:     method,
		Path:       path,
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
		StatusCode: statusCode,
		DurationMS: duration.Milliseconds(),
	}
	logEncoder.Encode(entry)
}
