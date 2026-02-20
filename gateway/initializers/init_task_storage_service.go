package initializers

import (
	pb "gateway/proto/taskpb"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var TaskStorageService pb.TaskServiceClient

func InitTaskStorageService() pb.TaskServiceClient {
	host := os.Getenv("TASK_STORAGE_SERVICE_HOST")
	if host == "" {
		log.Fatal("❌ Error: TASK_STORAGE_SERVICE_HOST not set")
	}
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Error: Failed to connect to gRPC server: %v", err)
	}
	log.Println("✅ Successfully connected to task storage service")
	TaskStorageService = pb.NewTaskServiceClient(conn)

	return TaskStorageService
}
