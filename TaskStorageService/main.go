package main

import (
	"TaskStorageService/initializers"
	"TaskStorageService/logger"
	"TaskStorageService/middleware"
	"TaskStorageService/models"
	"TaskStorageService/proto/server"
	"TaskStorageService/proto/taskpb"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		fmt.Println("gRPC-server start, port: ", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	}

	if initializers.KafkaProducer != nil {
		log.Println("Closing Kafka producer...")
		if err := initializers.KafkaProducer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}

	if initializers.RedisClient != nil {
		log.Println("Closing Redis connection...")
		if err := initializers.RedisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}

	log.Println("TaskStorageService shutdown complete")
}
