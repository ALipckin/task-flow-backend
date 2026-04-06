package app

import (
	"tasks/internal/config"
	"tasks/internal/domain/shard"
	"tasks/internal/infrastructure/adapters"
	"tasks/internal/ports"
	transportgrpc "tasks/internal/transport/grpc"
	grpcmiddleware "tasks/internal/transport/grpc/middleware"
	"tasks/internal/use_case"

	"google.golang.org/grpc"
)

// Container provides lazy dependency initialization for the tasks service.
// Each getter initializes its dependency once and reuses it on subsequent calls.
type Container struct {
	cfg      *config.Config
	shardMgr *shard.ShardManager

	repo      ports.Repository
	cache     ports.Cache
	producer  ports.EventProducer
	allocator ports.IDAllocator

	createTaskUC *use_case.CreateTask
	getTaskUC    *use_case.GetTask
	getTasksUC   *use_case.GetTasks
	deleteTaskUC *use_case.DeleteTask
	updateTaskUC *use_case.UpdateTask

	taskServer *transportgrpc.TaskServer
	grpcServer *grpc.Server
}

// NewContainer creates an empty container with lazy initialization.
func NewContainer() *Container {
	return &Container{}
}

// Config returns the application configuration.
func (c *Container) Config() *config.Config {
	if c.cfg == nil {
		c.cfg = config.AppConfig()
	}

	return c.cfg
}

// Infrastructure initializes shared infrastructure and returns the shard manager.
func (c *Container) Infrastructure() *shard.ShardManager {
	if c.shardMgr == nil {
		c.shardMgr = adapters.InitializeInfrastructure(c.Config())
	}

	return c.shardMgr
}

// Repository returns the task repository implementation.
func (c *Container) Repository() ports.Repository {
	if c.repo == nil {
		c.repo = adapters.NewPostgresRepository(c.Infrastructure())
	}

	return c.repo
}

// Cache returns the cache adapter implementation.
func (c *Container) Cache() ports.Cache {
	if c.cache == nil {
		_ = c.Infrastructure()
		c.cache = adapters.NewRedisCacheAdapter()
	}

	return c.cache
}

// Producer returns the event producer implementation.
func (c *Container) Producer() ports.EventProducer {
	if c.producer == nil {
		_ = c.Infrastructure()
		c.producer = adapters.NewKafkaProducerAdapter()
	}

	return c.producer
}

// Allocator returns the ID allocator implementation.
func (c *Container) Allocator() ports.IDAllocator {
	if c.allocator == nil {
		_ = c.Infrastructure()
		c.allocator = adapters.NewRedisIDAllocator()
	}

	return c.allocator
}

// CreateTaskUC returns CreateTask use-case.
func (c *Container) CreateTaskUC() *use_case.CreateTask {
	if c.createTaskUC == nil {
		c.createTaskUC = use_case.NewCreateTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
			c.Infrastructure(),
			c.Allocator(),
		)
	}

	return c.createTaskUC
}

// GetTaskUC returns GetTask use-case.
func (c *Container) GetTaskUC() *use_case.GetTask {
	if c.getTaskUC == nil {
		c.getTaskUC = use_case.NewGetTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
		)
	}

	return c.getTaskUC
}

// GetTasksUC returns GetTasks use-case.
func (c *Container) GetTasksUC() *use_case.GetTasks {
	if c.getTasksUC == nil {
		c.getTasksUC = use_case.NewGetTasks(
			c.Repository(),
			c.Infrastructure(),
			c.Allocator(),
		)
	}

	return c.getTasksUC
}

// DeleteTaskUC returns DeleteTask use-case.
func (c *Container) DeleteTaskUC() *use_case.DeleteTask {
	if c.deleteTaskUC == nil {
		c.deleteTaskUC = use_case.NewDeleteTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
		)
	}

	return c.deleteTaskUC
}

// UpdateTaskUC returns UpdateTask use-case.
func (c *Container) UpdateTaskUC() *use_case.UpdateTask {
	if c.updateTaskUC == nil {
		c.updateTaskUC = use_case.NewUpdateTask(
			c.Repository(),
			c.Producer(),
		)
	}

	return c.updateTaskUC
}

// TaskServer returns the gRPC task server with injected use-cases.
func (c *Container) TaskServer() *transportgrpc.TaskServer {
	if c.taskServer == nil {
		c.taskServer = &transportgrpc.TaskServer{
			CreateUC:   c.CreateTaskUC(),
			GetTaskUC:  c.GetTaskUC(),
			GetTasksUC: c.GetTasksUC(),
			DeleteUC:   c.DeleteTaskUC(),
			UpdateUC:   c.UpdateTaskUC(),
		}
	}

	return c.taskServer
}

// GRPCServer returns the configured gRPC server instance.
func (c *Container) GRPCServer() *grpc.Server {
	if c.grpcServer == nil {
		c.grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpcmiddleware.UnaryLoggingInterceptor()),
		)
	}

	return c.grpcServer
}

// Cleanup releases infrastructure resources.
func (c *Container) Cleanup() {
	adapters.CleanupInfrastructure()
}
