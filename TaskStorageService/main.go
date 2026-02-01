package main

import (
	"TaskStorageService/initializers"
	"TaskStorageService/logger"
	"TaskStorageService/middleware"
	"TaskStorageService/models"
	"TaskStorageService/proto/server"
	"TaskStorageService/proto/taskpb"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func init() {
	initializers.LoadEnvVariables()
	models.InitShardManager()
	initializers.ConnectRedis()
	initializers.InitProducer()
	logger.Init()
}

func main() {
	initializers.SyncDatabaseForShards()
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor()),
	)

	taskServer := &server.TaskServer{
		ShardManager: models.ShardMgr,
	}

	taskpb.RegisterTaskServiceServer(grpcServer, taskServer)
	port := ":" + os.Getenv("GRPC_PORT")
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Start error: %v", err)
	}

	fmt.Println("gRPC-server start, port: ", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
