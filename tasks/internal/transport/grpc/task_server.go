package grpc

import (
	"tasks/internal/application/ports/in/commands"
	ingrpc "tasks/internal/application/ports/in/grpc"
	"tasks/internal/application/ports/in/queries"
	"tasks/proto/taskpb"

	"google.golang.org/grpc"
)

var _ ingrpc.TaskService = (*TaskServer)(nil)

type TaskServer struct {
	taskpb.UnimplementedTaskServiceServer
	CreateUC   commands.CreateTaskHandler
	GetTaskUC  queries.GetTaskHandler
	GetTasksUC queries.GetTasksHandler
	DeleteUC   commands.DeleteTaskHandler
	UpdateUC   commands.UpdateTaskHandler
}

func (s *TaskServer) Register(reg grpc.ServiceRegistrar) {
	taskpb.RegisterTaskServiceServer(reg, s)
}
