package grpc

import (
	"context"
	"tasks/internal/use_case"
	"tasks/proto/taskpb"
)

// its have to be transport adapter
func (s *TaskServer) GetTasks(
	ctx context.Context,
	req *taskpb.GetTasksRequest,
) (*taskpb.GetTasksResponse, error) {

	cmd := use_case.GetTasksCommand{
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
		protoTasks = append(protoTasks, ToProto(task))
	}

	return &taskpb.GetTasksResponse{Tasks: protoTasks}, nil
}
