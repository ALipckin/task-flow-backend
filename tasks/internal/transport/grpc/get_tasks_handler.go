package grpc

import (
	"context"
	"tasks/internal/application/ports/in/queries"
	"tasks/proto/taskpb"
)

func (s *TaskServer) GetTasks(
	ctx context.Context,
	req *taskpb.GetTasksRequest,
) (*taskpb.GetTasksResponse, error) {

	cmd := queries.GetTasksQuery{
		Title:       req.Title,
		PerformerID: uint(req.PerformerId),
		CreatorID:   uint(req.CreatorId),
	}

	tasks, err := s.GetTasksUC.Execute(ctx, cmd)

	if err != nil {
		return nil, err
	}

	protoTasks := make([]*taskpb.Task, 0, len(tasks))
	for _, task := range tasks {
		protoTasks = append(protoTasks, ToProto(&task))
	}

	return &taskpb.GetTasksResponse{Tasks: protoTasks}, nil
}
