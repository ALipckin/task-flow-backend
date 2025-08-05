package helpers

import (
	"TaskStorageService/models"
	"TaskStorageService/proto/taskpb"
	"errors"
	"time"
)

func ApplyTaskFieldsFromRequest(task *models.Task, req interface{}) error {
	switch r := req.(type) {
	case *taskpb.CreateTaskRequest:
		task.Title = r.Title
		task.Description = r.Description
		task.PerformerId = uint(r.PerformerId)
		task.CreatorId = uint(r.CreatorId)
		task.Observers = models.ObserversFromIDs(r.ObserverIds)
		task.Status = r.Status
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
	case *taskpb.UpdateTaskRequest:
		task.Title = r.Title
		task.Description = r.Description
		task.PerformerId = uint(r.PerformerId)
		task.CreatorId = uint(r.CreatorId)
		task.Observers = models.ObserversFromIDs(r.ObserverIds)
		task.Status = r.Status
		task.UpdatedAt = time.Now()
	default:
		return errors.New("unknown request type")
	}
	return nil
}
