package grpc

import (
	"context"
	"errors"
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/out"
	"tasks/internal/domain"
	"tasks/proto/taskpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *TaskServer) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.TaskResponse, error) {
	cmd := commands.UpdateTaskCommand{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		PerformerID: uint(req.PerformerId),
		CreatorID:   uint(req.CreatorId),
		ObserverIDs: req.ObserverIds,
	}

	task, err := s.UpdateUC.Execute(ctx, cmd)
	if err != nil {
		if errors.Is(err, out.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
		}
		if domainErr, ok := mapDomainError(err); ok {
			return nil, domainErr
		}
		return nil, err
	}

	return &taskpb.TaskResponse{Task: ToProto(&task)}, nil
}

func mapDomainError(err error) (error, bool) {
	switch {
	case errors.Is(err, domain.ErrEmptyTitle),
		errors.Is(err, domain.ErrTitleTooLong),
		errors.Is(err, domain.ErrDescriptionTooLong),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrInvalidPerformer),
		errors.Is(err, domain.ErrInvalidCreator),
		errors.Is(err, domain.ErrObserverIsPerformer),
		errors.Is(err, domain.ErrDuplicateObserver),
		errors.Is(err, domain.ErrTooManyObservers),
		errors.Is(err, domain.ErrTaskDeleted):
		return status.Error(codes.InvalidArgument, err.Error()), true
	default:
		return nil, false
	}
}
