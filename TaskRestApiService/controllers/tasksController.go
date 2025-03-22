package controllers

import (
	"TaskRestApiService/initializers"
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
		initializers.LogToKafka("error", "TasksCreate", "Failed to bind JSON", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := tc.GRPCClient.CreateTask(context.Background(), &req)
	if err != nil {
		initializers.LogToKafka("error", "TasksCreate", "Failed to create task", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksCreate", "Task successfully created", resp)
	c.JSON(http.StatusCreated, resp)
}

func (tc *TaskController) TasksIndex(c *gin.Context) {
	resp, err := tc.GRPCClient.GetTasks(context.Background(), &pb.GetTasksRequest{})
	if err != nil {
		initializers.LogToKafka("error", "TasksIndex", "Failed to retrieve task list", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksIndex", "Task list retrieved successfully", resp.Tasks)
	c.JSON(http.StatusOK, resp.Tasks)
}

func (tc *TaskController) TasksShow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		initializers.LogToKafka("error", "TasksShow", "Invalid task ID format", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	resp, err := tc.GRPCClient.GetTask(context.Background(), &pb.GetTaskRequest{Id: id})
	if err != nil {
		initializers.LogToKafka("error", "TasksShow", "Task not found", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksShow", "Task retrieved successfully", resp)
	c.JSON(http.StatusOK, resp)
}

func (tc *TaskController) TasksUpdate(c *gin.Context) {
	var req pb.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		initializers.LogToKafka("error", "TasksUpdate", "Failed to bind JSON", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := tc.GRPCClient.UpdateTask(context.Background(), &req)
	if err != nil {
		initializers.LogToKafka("error", "TasksUpdate", "Failed to update task", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksUpdate", "Task updated successfully", resp)
	c.JSON(http.StatusOK, resp)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	initializers.LogToKafka("info", "TasksDelete", "Task deleted successfully", id)
	c.JSON(http.StatusNoContent, nil)
}
