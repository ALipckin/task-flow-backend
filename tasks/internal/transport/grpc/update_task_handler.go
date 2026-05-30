package grpc

import (
	"context"
	"errors"
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/out"
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
		return nil, err
	}

	return &taskpb.TaskResponse{Task: ToProto(&task)}, nil
}
