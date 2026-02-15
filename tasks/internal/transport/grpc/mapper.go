package grpc

import (
	"tasks/internal/domain"
	"tasks/internal/infrastructure/persistence"
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
		ObserverIds: observersToIDs(task.Observers),
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
		Observers:   idsToObservers(pb.ObserverIds),
		Status:      pb.Status,
		CreatedAt:   timestampToTime(pb.CreatedAt),
		UpdatedAt:   timestampToTime(pb.UpdatedAt),
	}
}

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

func observersToIDs(observers []persistence.Observer) []uint64 {
	if len(observers) == 0 {
		return nil
	}

	ids := make([]uint64, len(observers))
	for i, o := range observers {
		ids[i] = uint64(o.UserId) // предполагаем, что там UserID
	}

	return ids
}

func idsToObservers(ids []uint64) []persistence.Observer {
	if len(ids) == 0 {
		return nil
	}

	observers := make([]persistence.Observer, len(ids))
	for i, id := range ids {
		observers[i] = persistence.Observer{
			UserId: uint(id),
		}
	}

	return observers
}

func timestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
