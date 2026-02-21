package grpc

import (
	"tasks/internal/use_case"
	"tasks/proto/taskpb"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	CreateUC   *use_case.CreateTask
	GetTaskUC  *use_case.GetTask
	GetTasksUC *use_case.GetTasks
	DeleteUC   *use_case.DeleteTask
	UpdateUC   *use_case.UpdateTask
}
