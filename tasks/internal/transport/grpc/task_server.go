package grpc

import (
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/in/queries"
	"tasks/proto/taskpb"
)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	CreateUC   commands.CreateTaskHandler
	GetTaskUC  queries.GetTaskHandler
	GetTasksUC queries.GetTasksHandler
	DeleteUC   commands.DeleteTaskHandler
	UpdateUC   commands.UpdateTaskHandler
}
