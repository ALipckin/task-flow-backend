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

// Create Task
// @Summary      Create a new task
// @Description  Creates a task with title, description, creator, performer, etc.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body      pb.CreateTaskRequest  true  "Task input"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks [post]
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

// List Tasks
// @Summary      Get list of tasks
// @Description  Retrieves a list of tasks based on optional filters
// @Tags         tasks
// @Produce      json
// @Param        title        query     string  false  "Filter by title"
// @Param        creator_id   query     uint64  false  "Filter by creator ID"
// @Param        performer_id query     uint64  false  "Filter by performer ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks [get]
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

// Get Task by ID
// @Summary      Get task by ID
// @Description  Retrieves a single task by its ID
// @Tags         tasks
// @Produce      json
// @Param        id   path      uint64  true  "Task ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id} [get]
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

// Update Task
// @Summary      Update a task
// @Description  Updates an existing task with given data
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path      uint64                 true  "Task ID"
// @Param        task  body      pb.UpdateTaskRequest   true  "Updated task input"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id} [put]
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

// Delete Task
// @Summary      Delete a task
// @Description  Deletes a task by ID
// @Tags         tasks
// @Produce      json
// @Param        id   path      uint64  true  "Task ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id} [delete]
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
