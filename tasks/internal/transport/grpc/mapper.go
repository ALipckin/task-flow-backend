package grpc

import (
	"tasks/internal/domain"
	"tasks/proto/taskpb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProto(task *domain.Task) *taskpb.Task {
	if task == nil {
		return nil
	}

	return &taskpb.Task{
		Id:          uint64(task.ID),
		Title:       task.Title,
		Description: task.Description,
		PerformerId: uint64(task.PerformerId),
		CreatorId:   uint64(task.CreatorId),
		ObserverIds: task.ObserverUserIDs(),
		Status:      task.Status.String(),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}

func uint64SliceToUint(src []uint64) []uint {
	if len(src) == 0 {
		return nil
	}
	res := make([]uint, len(src))
	for i, v := range src {
		res[i] = uint(v)
	}
	return res
}
