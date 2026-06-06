package domain

import "strings"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

func (s TaskStatus) String() string {
	return string(s)
}

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusDone, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (s TaskStatus) CanTransitionTo(next TaskStatus) bool {
	if s == next {
		return true
	}

	switch s {
	case TaskStatusPending:
		return next == TaskStatusInProgress || next == TaskStatusCancelled
	case TaskStatusInProgress:
		return next == TaskStatusDone || next == TaskStatusCancelled
	case TaskStatusDone, TaskStatusCancelled:
		return false
	default:
		return false
	}
}

func ParseTaskStatus(raw string) (TaskStatus, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return TaskStatusPending, nil
	}
	if normalized == "new" {
		normalized = string(TaskStatusPending)
	}

	status := TaskStatus(normalized)
	if !status.IsValid() {
		return "", ErrInvalidStatus
	}

	return status, nil
}
