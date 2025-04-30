package main

import (
	"TaskStorageService/initializers"
	"TaskStorageService/logger"
	"TaskStorageService/middleware"
	"TaskStorageService/models"
	"TaskStorageService/proto/server"
	"TaskStorageService/proto/taskpb"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

func init() {
	initializers.LoadEnvVariables()
	models.ConnectToDB()
	initializers.ConnectRedis()
	initializers.InitProducer()
	logger.Init()
}

func main() {
	initializers.SyncDatabase(models.DB)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor()),
	)

	taskServer := &server.TaskServer{
		DB: models.DB,
	}

	taskpb.RegisterTaskServiceServer(grpcServer, taskServer)
	port := ":" + os.Getenv("GRPC_PORT")
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Start error: %v", err)
	}

	fmt.Println("gRPC-server start, port: ", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка запуска gRPC: %v", err)
	}
}
