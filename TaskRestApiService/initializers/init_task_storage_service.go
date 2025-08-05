package initializers

import (
	pb "TaskRestApiService/proto/taskpb"
	"google.golang.org/grpc"
	"log"
	"os"
)

var TaskStorageService pb.TaskServiceClient

func InitTaskStorageService() pb.TaskServiceClient {
	host := os.Getenv("TASK_STORAGE_SERVICE_HOST")
	if host == "" {
		log.Fatal("❌ Error: TASK_STORAGE_SERVICE_HOST not set")
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Error: Failed to connect to gRPC server: %v", err)
	}
	log.Println("✅ Successfully connected to task storage service")
	TaskStorageService = pb.NewTaskServiceClient(conn)

	return TaskStorageService
}
