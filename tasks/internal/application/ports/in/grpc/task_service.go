package grpc

import "google.golang.org/grpc"

// TaskService is the inbound port for exposing task operations over gRPC.
// Transport adapters implement this interface; the composition root registers it.
type TaskService interface {
	Register(reg grpc.ServiceRegistrar)
}
