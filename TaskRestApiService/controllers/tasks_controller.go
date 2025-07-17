package controllers

import (
	"TaskRestApiService/logger"
	pb "TaskRestApiService/proto/taskpb"
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
		logger.Log(logger.LevelError, "Failed to bind JSON", gin.H{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := tc.GRPCClient.CreateTask(context.Background(), &req)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to create task", gin.H{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logger.Log(logger.LevelInfo, "Task successfully created", gin.H{"task": resp})
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (tc *TaskController) TasksIndex(c *gin.Context) {
	title := c.Query("title")
	creatorID := c.Query("creator_id")
	performerID := c.Query("performer_id")

	req := &pb.GetTasksRequest{}

	if title != "" {
		req.Title = title
	}

	if creatorID != "" {
		creatorID, err := strconv.ParseUint(creatorID, 10, 64)
		if err != nil {
			logger.Log(logger.LevelError, "Invalid creator_id format", gin.H{"creator_id": creatorID})
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid creator_id"})
			return
		}
		req.CreatorId = creatorID
	}

	if performerID != "" {
		performerID, err := strconv.ParseUint(performerID, 10, 64)
		if err != nil {
			logger.Log(logger.LevelError, "Invalid performer_id format", gin.H{"performer_id": performerID})
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid performer_id"})
			return
		}
		req.PerformerId = performerID
	}

	resp, err := tc.GRPCClient.GetTasks(context.Background(), req)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to retrieve task list", gin.H{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logger.Log(logger.LevelInfo, "Task list retrieved successfully", gin.H{"tasks": resp.Tasks})
	c.JSON(http.StatusOK, gin.H{"data": resp.Tasks})
}

func (tc *TaskController) TasksShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Log(logger.LevelError, "Invalid task ID format", gin.H{"task_id": idStr})
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid task ID"})
		return
	}

	resp, err := tc.GRPCClient.GetTask(context.Background(), &pb.GetTaskRequest{Id: id})
	if err != nil {
		logger.Log(logger.LevelError, "Task not found", gin.H{"error": err.Error()})
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	logger.Log(logger.LevelInfo, "Task retrieved successfully", gin.H{"task": resp.Task})
	c.JSON(http.StatusOK, gin.H{"data": resp.Task})
}

func (tc *TaskController) TasksUpdate(c *gin.Context) {
	var req pb.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log(logger.LevelError, "Failed to bind JSON", gin.H{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := tc.GRPCClient.UpdateTask(context.Background(), &req)
	if err != nil {
		logger.Log(logger.LevelError, "Failed to update task", gin.H{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logger.Log(logger.LevelInfo, "Task updated successfully", gin.H{"task": resp})
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (tc *TaskController) TasksDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Log(logger.LevelError, "Invalid task ID format", gin.H{"task_id": idStr})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	_, err = tc.GRPCClient.DeleteTask(context.Background(), &pb.DeleteTaskRequest{Id: id})
	if err != nil {
		logger.Log(logger.LevelError, "Failed to delete task", gin.H{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	logger.Log(logger.LevelInfo, "Task deleted successfully", gin.H{"task_id": id})
	c.JSON(http.StatusNoContent, nil)
}
