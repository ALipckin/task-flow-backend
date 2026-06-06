package grpc

import (
	"errors"
	"tasks/internal/domain"
	"tasks/proto/taskpb"
	"time"
)

func ApplyTaskFieldsFromRequest(task *domain.Task, req interface{}) error {
	switch r := req.(type) {
	case *taskpb.CreateTaskRequest:
		task.Title = r.Title
		task.Description = r.Description
		task.PerformerId = uint(r.PerformerId)
		task.CreatorId = uint(r.CreatorId)
		task.Observers = domain.ObserversFromUserIDs(r.ObserverIds)
		if status, err := domain.ParseTaskStatus(r.Status); err == nil {
			task.Status = status
		}
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
	case *taskpb.UpdateTaskRequest:
		task.Title = r.Title
		task.Description = r.Description
		task.PerformerId = uint(r.PerformerId)
		task.CreatorId = uint(r.CreatorId)
		task.Observers = domain.ObserversFromUserIDs(r.ObserverIds)
		if status, err := domain.ParseTaskStatus(r.Status); err == nil {
			task.Status = status
		}
		task.UpdatedAt = time.Now()
	default:
		return errors.New("unknown request type")
	}
	return nil
}
