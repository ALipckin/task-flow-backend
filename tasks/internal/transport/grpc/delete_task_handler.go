package grpc

import (
	"context"
	"tasks/internal/use_case"
	"tasks/proto/taskpb"
)

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	cmd := use_case.DeleteTaskCommand{
		ID: req.Id,
	}
	ok, err := s.DeleteUC.Execute(ctx, cmd)

	if err != nil {
		return nil, err
	}

	if !ok {
		return &taskpb.DeleteTaskResponse{Message: "Task not found"}, nil
	}

	return &taskpb.DeleteTaskResponse{Message: "Task deleted"}, nil
}
