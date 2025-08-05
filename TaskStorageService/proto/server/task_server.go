package server

import (
	"TaskStorageService/helpers"
	"TaskStorageService/initializers"
	"TaskStorageService/models"
	"TaskStorageService/proto/taskpb"
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	DB *gorm.DB
}

func (s *TaskServer) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskResponse, error) {
	var task models.Task
	if err := helpers.ApplyTaskFieldsFromRequest(&task, req); err != nil {
		return nil, err
	}

	if err := s.DB.Create(&task).Error; err != nil {
		return nil, err
	}

	if err := helpers.CacheSetTask(ctx, task); err != nil {
		return nil, err
	}

	if err := helpers.SendTaskEventToKafka("TaskCreated", task); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task)}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	task, err := helpers.CacheGetTask(ctx, uint(req.Id))
	if err == nil {
		return &taskpb.TaskResponse{Task: convertToProto(*task)}, nil
	}

	task = &models.Task{}
	if err := s.DB.First(task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := helpers.CacheSetTask(ctx, *task); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(*task)}, nil
}

func (s *TaskServer) GetTasks(ctx context.Context, req *taskpb.GetTasksRequest) (*taskpb.GetTasksResponse, error) {
	var tasks []models.Task
	query := s.DB

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
		return nil, err
	}

	protoTasks := make([]*taskpb.Task, 0, len(tasks))
	for _, t := range tasks {
		protoTasks = append(protoTasks, convertToProto(t))
	}

	return &taskpb.GetTasksResponse{Tasks: protoTasks}, nil
}

func (s *TaskServer) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.TaskResponse, error) {
	var task models.Task
	if err := s.DB.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	// Очистим старых наблюдателей
	if err := s.DB.Where("task_id = ?", task.ID).Delete(&models.Observer{}).Error; err != nil {
		return nil, err
	}

	if err := helpers.ApplyTaskFieldsFromRequest(&task, req); err != nil {
		return nil, err
	}

	if err := s.DB.Save(&task).Error; err != nil {
		return nil, err
	}

	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))

	if err := helpers.SendTaskEventToKafka("TaskUpdated", task); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task)}, nil
}

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	var task models.Task
	if err := s.DB.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Delete(&task).Error; err != nil {
		return nil, err
	}

	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))

	if err := helpers.SendTaskEventToKafka("TaskDeleted", task); err != nil {
		return nil, err
	}

	return &taskpb.DeleteTaskResponse{Message: "Task deleted"}, nil
}

func convertToProto(task models.Task) *taskpb.Task {
	return &taskpb.Task{
		Id:          uint64(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		PerformerId: uint64(task.PerformerId),
		CreatorId:   uint64(task.CreatorId),
		ObserverIds: task.ObserverIDs(),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}
