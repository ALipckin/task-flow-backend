package grpc

import (
	"tasks/internal/infrastructure/persistence"
	"tasks/proto/taskpb"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func convertToProto(task persistence.Task, shard *gorm.DB) *taskpb.Task {
	return &taskpb.Task{
		Id:          uint64(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		PerformerId: uint64(task.PerformerId),
		CreatorId:   uint64(task.CreatorId),
		ObserverIds: task.ObserverIDs(shard),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}
