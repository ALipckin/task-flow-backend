package controllers

import (
	"TaskRestApiService/initializers"
	pb "TaskRestApiService/proto/taskpb"
	"TaskRestApiService/services"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type TaskController struct {
	GRPCClient pb.TaskServiceClient
}

func NewTaskController(grpcClient pb.TaskServiceClient) *TaskController {
	return &TaskController{GRPCClient: grpcClient}
}

func (tc *TaskController) TasksCreate(c *gin.Context) {
	var req pb.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		initializers.LogToKafka("error", "TasksCreate", "Failed to bind JSON", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := tc.GRPCClient.CreateTask(context.Background(), &req)
	if err != nil {
		initializers.LogToKafka("error", "TasksCreate", "Failed to create task", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksCreate", "Task successfully created", resp)
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (tc *TaskController) TasksIndex(c *gin.Context) {
	resp, err := tc.GRPCClient.GetTasks(context.Background(), &pb.GetTasksRequest{})
	if err != nil {
		initializers.LogToKafka("error", "TasksIndex", "Failed to retrieve task list", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	tasksWithUsers, err := tc.enrichTasksWithUserData(resp.Tasks)
	if err != nil {
		initializers.LogToKafka("error", "TasksIndex", "Failed to get users data", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get users data"})
		return
	}

	initializers.LogToKafka("info", "TasksIndex", "Task list with users retrieved successfully", tasksWithUsers)
	c.JSON(http.StatusOK, gin.H{"data": tasksWithUsers})
}

func (tc *TaskController) TasksShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		initializers.LogToKafka("error", "TasksShow", "Invalid task ID format", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid task ID"})
		return
	}

	resp, err := tc.GRPCClient.GetTask(context.Background(), &pb.GetTaskRequest{Id: id})
	if err != nil {
		initializers.LogToKafka("error", "TasksShow", "Task not found", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	taskWithUsers, err := tc.enrichTasksWithUserData([]*pb.Task{resp.Task})
	if err != nil {
		initializers.LogToKafka("error", "TasksShow", "Failed to get users data", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get users data"})
		return
	}

	initializers.LogToKafka("info", "TasksShow", "Task with user data retrieved successfully", taskWithUsers[0])
	c.JSON(http.StatusOK, gin.H{"data": taskWithUsers[0]})
}

func (tc *TaskController) enrichTasksWithUserData(tasks []*pb.Task) ([]gin.H, error) {
	userIDSet := make(map[int]struct{})
	for _, task := range tasks {
		userIDSet[int(task.PerformerId)] = struct{}{}
		userIDSet[int(task.CreatorId)] = struct{}{}
		for _, observerId := range task.ObserverIds {
			userIDSet[int(observerId)] = struct{}{}
		}
	}

	userIDs := make([]int, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	users, err := services.GetUsersData(userIDs)
	if err != nil {
		return nil, err
	}

	userMap := make(map[int]string)
	for _, user := range users {
		userMap[user.ID] = user.Email
	}

	tasksWithUsers := make([]gin.H, len(tasks))
	for i, task := range tasks {
		tasksWithUsers[i] = tc.enrichTaskWithUserData(task, userMap)
	}

	return tasksWithUsers, nil
}

func (tc *TaskController) enrichTaskWithUserData(task *pb.Task, userMap map[int]string) gin.H {
	return gin.H{
		"id":          task.Id,
		"title":       task.Title,
		"description": task.Description,
		"performer": gin.H{
			"id":    task.PerformerId,
			"email": userMap[int(task.PerformerId)],
		},
		"creator": gin.H{
			"id":    task.CreatorId,
			"email": userMap[int(task.CreatorId)],
		},
		"observers": func() []gin.H {
			observers := make([]gin.H, len(task.ObserverIds))
			for j, observerId := range task.ObserverIds {
				observers[j] = gin.H{
					"id":    observerId,
					"email": userMap[int(observerId)],
				}
			}
			return observers
		}(),
		"status":     task.Status,
		"created_at": task.CreatedAt,
		"updated_at": task.UpdatedAt,
	}
}

func (tc *TaskController) TasksUpdate(c *gin.Context) {
	var req pb.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		initializers.LogToKafka("error", "TasksUpdate", "Failed to bind JSON", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := tc.GRPCClient.UpdateTask(context.Background(), &req)
	if err != nil {
		initializers.LogToKafka("error", "TasksUpdate", "Failed to update task", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksUpdate", "Task updated successfully", resp)
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (tc *TaskController) TasksDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		initializers.LogToKafka("error", "TasksDelete", "Invalid task ID format", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	_, err = tc.GRPCClient.DeleteTask(context.Background(), &pb.DeleteTaskRequest{Id: id})
	if err != nil {
		initializers.LogToKafka("error", "TasksDelete", "Failed to delete task", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksDelete", "Task deleted successfully", id)
	c.JSON(http.StatusNoContent, nil)
}
