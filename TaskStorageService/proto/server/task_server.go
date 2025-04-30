package server

import (
	"TaskStorageService/initializers"
	"TaskStorageService/models"
	"TaskStorageService/proto/taskpb"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"time"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	DB *gorm.DB
}

func getTaskMessage(eventName string, task models.Task) []byte {
	message := map[string]interface{}{
		"event":         eventName,
		"task_id":       task.ID,
		"title":         task.Title,
		"description":   task.Description,
		"performer_id":  task.PerformerId,
		"creator_id":    task.CreatorId,
		"observers_ids": task.ObserverIDs(),
		"status":        task.Status,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}
	messageJSON, _ := json.Marshal(message)

	return messageJSON
}

func (s *TaskServer) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskResponse, error) {
	task := models.Task{
		Title:       req.Title,
		Description: req.Description,
		PerformerId: uint(req.PerformerId),
		CreatorId:   uint(req.CreatorId),
		Observers:   models.ObserversFromIDs(req.ObserverIds),
		Status:      req.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.DB.Create(&task).Error; err != nil {
		return nil, err
	}

	redisKey := fmt.Sprintf("task:%d", task.ID)
	taskJSON, _ := json.Marshal(task)
	initializers.RedisClient.Set(ctx, redisKey, taskJSON, 10*time.Minute)

	err := initializers.SendMessageToKafka(getTaskMessage("TaskCreated", task))
	if err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task)}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	redisKey := fmt.Sprintf("task:%d", req.Id)

	taskJSON, err := initializers.RedisClient.Get(ctx, redisKey).Result()
	if err == nil {
		var task models.Task
		json.Unmarshal([]byte(taskJSON), &task)
		return &taskpb.TaskResponse{Task: convertToProto(task)}, nil
	}

	var task models.Task
	if err := s.DB.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	taskJSONBytes, _ := json.Marshal(task)
	taskJSON = string(taskJSONBytes)
	initializers.RedisClient.Set(ctx, redisKey, taskJSON, 10*time.Minute)

	return &taskpb.TaskResponse{Task: convertToProto(task)}, nil
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

	var protoTasks []*taskpb.Task
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
	if err := s.DB.Where("task_id = ?", task.ID).Delete(&models.Observer{}).Error; err != nil {
		return nil, err
	}

	task.Title = req.Title
	task.Description = req.Description
	task.PerformerId = uint(req.PerformerId)
	task.CreatorId = uint(req.CreatorId)
	task.Observers = models.ObserversFromIDs(req.ObserverIds)
	task.Status = req.Status
	task.UpdatedAt = time.Now()

	if err := s.DB.Save(&task).Error; err != nil {
		return nil, err
	}

	redisKey := fmt.Sprintf("task:%d", task.ID)
	initializers.RedisClient.Del(ctx, redisKey)

	err := initializers.SendMessageToKafka(getTaskMessage("TaskUpdated", task))
	if err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task)}, nil
}

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	var task models.Task
	if err := s.DB.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Delete(&models.Task{}, req.Id).Error; err != nil {
		return nil, err
	}

	redisKey := fmt.Sprintf("task:%d", req.Id)
	initializers.RedisClient.Del(ctx, redisKey)

	err := initializers.SendMessageToKafka(getTaskMessage("TaskDeleted", task))
	if err != nil {
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
