package grpc

import (
	"tasks/internal/domain"
	"tasks/proto/taskpb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProto(domainTask domain.Task) *taskpb.Task {
	return &taskpb.Task{
		Id:          uint64(domainTask.ID),
		Title:       domainTask.Title,
		Description: domainTask.Description,
		Status:      domainTask.Status,
		PerformerId: uint64(domainTask.PerformerId),
		CreatorId:   uint64(domainTask.CreatorId),
		ObserverIds: nil,
		CreatedAt:   timestamppb.New(domainTask.CreatedAt),
		UpdatedAt:   timestamppb.New(domainTask.UpdatedAt),
	}
}
