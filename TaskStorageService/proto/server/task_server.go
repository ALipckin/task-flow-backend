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
	ShardManager *models.ShardManager
}

func (s *TaskServer) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskResponse, error) {
	var task models.Task
	if err := helpers.ApplyTaskFieldsFromRequest(&task, req); err != nil {
		return nil, err
	}

	shardIndex := s.ShardManager.NextShardIndex()
	id, err := s.ShardManager.AllocNextID(shardIndex)
	if err != nil {
		return nil, err
	}

	task.ID = id
	shard := s.ShardManager.GetShardByIndex(shardIndex)
	if err := shard.Create(&task).Error; err != nil {
		return nil, err
	}

	if err := helpers.CacheSetTask(ctx, task); err != nil {
		return nil, err
	}

	if err := helpers.SendTaskEventToKafka("TaskCreated", task, shard); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task, shard)}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	taskID := uint(req.Id)
	task, err := helpers.CacheGetTask(ctx, taskID)
	if err == nil {
		shard := s.ShardManager.GetShard(taskID)
		return &taskpb.TaskResponse{Task: convertToProto(*task, shard)}, nil
	}

	shard := s.ShardManager.GetShard(taskID)
	task = &models.Task{}
	if err := shard.First(task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := helpers.CacheSetTask(ctx, *task); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(*task, shard)}, nil
}

func (s *TaskServer) GetTasks(ctx context.Context, req *taskpb.GetTasksRequest) (*taskpb.GetTasksResponse, error) {
	allShards := s.ShardManager.GetAllShards()
	allTasks := make([]models.Task, 0)

	for _, shard := range allShards {
		var tasks []models.Task
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

		allTasks = append(allTasks, tasks...)
	}

	protoTasks := make([]*taskpb.Task, 0, len(allTasks))
	for _, t := range allTasks {
		shard := s.ShardManager.GetShard(t.ID)
		protoTasks = append(protoTasks, convertToProto(t, shard))
	}

	return &taskpb.GetTasksResponse{Tasks: protoTasks}, nil
}

func (s *TaskServer) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.TaskResponse, error) {
	taskID := uint(req.Id)
	shard := s.ShardManager.GetShard(taskID)

	var task models.Task
	if err := shard.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := shard.Where("task_id = ?", task.ID).Delete(&models.Observer{}).Error; err != nil {
		return nil, err
	}

	if err := helpers.ApplyTaskFieldsFromRequest(&task, req); err != nil {
		return nil, err
	}

	if err := shard.Save(&task).Error; err != nil {
		return nil, err
	}

	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))

	if err := helpers.SendTaskEventToKafka("TaskUpdated", task, shard); err != nil {
		return nil, err
	}

	return &taskpb.TaskResponse{Task: convertToProto(task, shard)}, nil
}

func (s *TaskServer) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*taskpb.DeleteTaskResponse, error) {
	taskID := uint(req.Id)
	shard := s.ShardManager.GetShard(taskID)

	var task models.Task
	if err := shard.First(&task, req.Id).Error; err != nil {
		return nil, err
	}

	if err := shard.Delete(&task).Error; err != nil {
		return nil, err
	}

	initializers.RedisClient.Del(ctx, helpers.CacheKey(task.ID))

	if err := helpers.SendTaskEventToKafka("TaskDeleted", task, shard); err != nil {
		return nil, err
	}

	return &taskpb.DeleteTaskResponse{Message: "Task deleted"}, nil
}

func convertToProto(task models.Task, shard *gorm.DB) *taskpb.Task {
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
