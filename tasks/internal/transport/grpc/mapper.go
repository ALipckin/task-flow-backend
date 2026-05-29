package grpc

import (
	"tasks/internal/domain"
	"tasks/proto/taskpb"
	"time"

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
		Status:      task.Status,
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}

func ToDomain(pb *taskpb.Task) *domain.Task {
	if pb == nil {
		return nil
	}

	return &domain.Task{
		ID:          uint(pb.Id),
		Title:       pb.Title,
		Description: pb.Description,
		PerformerId: uint(pb.PerformerId),
		CreatorId:   uint(pb.CreatorId),
		Observers:   domain.ObserversFromUserIDs(pb.ObserverIds),
		Status:      pb.Status,
		CreatedAt:   timestampToTime(pb.CreatedAt),
		UpdatedAt:   timestampToTime(pb.UpdatedAt),
	}
}

//nolint:unused // Reserved conversion helper for partial update payload mapping.
func uintSliceToUint64(src []uint) []uint64 {
	if len(src) == 0 {
		return nil
	}
	res := make([]uint64, len(src))
	for i, v := range src {
		res[i] = uint64(v)
	}
	return res
}

//nolint:unused // Reserved conversion helper for partial update payload mapping.
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

func timestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
