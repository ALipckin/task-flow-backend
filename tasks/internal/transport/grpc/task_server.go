package grpc

import (
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/in/queries"
	"tasks/proto/taskpb"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	CreateUC   *commands.CreateTask
	GetTaskUC  *queries.GetTask
	GetTasksUC *queries.GetTasks
	DeleteUC   *commands.DeleteTask
	UpdateUC   *commands.UpdateTask
}
