package grpc

import (
	"tasks/internal/domain"
	"tasks/proto/taskpb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

//nolint:unused // Kept as legacy mapper used by external branch integrations.
func convertToProto(task domain.Task) *taskpb.Task {
	return &taskpb.Task{
		Id:          uint64(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		PerformerId: uint64(task.PerformerId),
		CreatorId:   uint64(task.CreatorId),
		ObserverIds: task.ObserverUserIDs(),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}
