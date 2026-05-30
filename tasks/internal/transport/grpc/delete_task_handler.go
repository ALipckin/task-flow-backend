package grpc

import (
	"context"
	"tasks/internal/application/ports/in/commands"
	"tasks/proto/taskpb"
)

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	cmd := commands.DeleteTaskCommand{
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
