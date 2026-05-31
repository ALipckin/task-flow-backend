package grpc

import (
	"context"
	"tasks/internal/application/ports/in/queries"
	"tasks/proto/taskpb"
)

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	cmd := queries.GetTaskQuery{
		ID: req.Id,
	}
	task, err := s.GetTaskUC.Execute(ctx, cmd)

	if err != nil {
		return nil, err
	}
	toProtoTask := ToProto(&task)
	return &taskpb.TaskResponse{Task: toProtoTask}, nil
}
