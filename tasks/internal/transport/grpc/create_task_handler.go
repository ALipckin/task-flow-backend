package grpc

import (
	"context"
	"tasks/internal/use_case"
	"tasks/proto/taskpb"
)

func (s *TaskServer) CreateTask(
	ctx context.Context,
	req *taskpb.CreateTaskRequest,
) (*taskpb.TaskResponse, error) {

	cmd := use_case.CreateTaskCommand{
		Title:       req.Title,
		Description: req.Description,
		PerformerID: uint(req.PerformerId),
		CreatorID:   uint(req.CreatorId),
	}

	task, err := s.CreateUC.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{
		Task: ToProto(&task),
	}, nil
}
