package server

import (
	"context"
	"tasks/internal/infrastructure/persistence"
	"tasks/proto/taskpb"
)

func (s *TaskServer) GetTasks(ctx context.Context, req *taskpb.GetTasksRequest) (*taskpb.GetTasksResponse, error) {
	allShards := s.ShardManager.GetAllShards()
	protoTasks := make([]*taskpb.Task, 0)

	for _, shard := range allShards {
		var tasks []persistence.Task
		query := shard

		if req.Title != "" {
			query = query.Where("title = ?", req.Title)
		}
		if req.CreatorId != 0 {
			query = query.Where("creator_id = ?", req.CreatorId)
		}
		if req.PerformerId != 0 {
			query = query.Where("performer_id = ?", req.PerformerId)
		}

		if err := query.Find(&tasks).Error; err != nil {
			continue
		}
		for _, t := range tasks {
			protoTasks = append(protoTasks, convertToProto(t, shard))
		}
	}

	return &taskpb.GetTasksResponse{Tasks: protoTasks}, nil
}
