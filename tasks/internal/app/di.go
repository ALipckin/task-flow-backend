package app

import (
	"tasks/internal/application/ports/in/commands"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/application/services"
	"tasks/internal/config"
	"tasks/internal/infrastructure/adapters"
	"tasks/internal/infrastructure/sharding/shard"
	transportgrpc "tasks/internal/transport/grpc"
	grpcmiddleware "tasks/internal/transport/grpc/middleware"

	"google.golang.org/grpc"
)

// Container provides lazy dependency initialization for the tasks service.
// Each getter initializes its dependency once and reuses it on subsequent calls.
type Container struct {
	cfg      *config.Config
	shardMgr *shard.ShardManager

	repo      out.Repository
	cache     out.Cache
	producer  out.EventProducer
	allocator out.IDAllocator

	createTaskUC *services.CreateTask
	getTaskUC    *services.GetTask
	getTasksUC   *services.GetTasks
	deleteTaskUC *services.DeleteTask
	updateTaskUC *services.UpdateTask

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
func (c *Container) Repository() out.Repository {
	if c.repo == nil {
		c.repo = adapters.NewPostgresRepository(c.Infrastructure())
	}

	return c.repo
}

// Cache returns the cache adapter implementation.
func (c *Container) Cache() out.Cache {
	if c.cache == nil {
		_ = c.Infrastructure()
		c.cache = adapters.NewRedisCacheAdapter()
	}

	return c.cache
}

// Producer returns the event producer implementation.
func (c *Container) Producer() out.EventProducer {
	if c.producer == nil {
		_ = c.Infrastructure()
		c.producer = adapters.NewKafkaProducerAdapter()
	}

	return c.producer
}

// Allocator returns the ID allocator implementation.
func (c *Container) Allocator() out.IDAllocator {
	if c.allocator == nil {
		_ = c.Infrastructure()
		c.allocator = adapters.NewRedisIDAllocator()
	}

	return c.allocator
}

// CreateTaskUC returns CreateTask use-case.
func (c *Container) CreateTaskUC() commands.CreateTaskHandler {
	if c.createTaskUC == nil {
		c.createTaskUC = services.NewCreateTask(
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
func (c *Container) GetTaskUC() queries.GetTaskHandler {
	if c.getTaskUC == nil {
		c.getTaskUC = services.NewGetTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
		)
	}

	return c.getTaskUC
}

// GetTasksUC returns GetTasks use-case.
func (c *Container) GetTasksUC() queries.GetTasksHandler {
	if c.getTasksUC == nil {
		c.getTasksUC = services.NewGetTasks(
			c.Repository(),
			c.Infrastructure(),
			c.Allocator(),
		)
	}

	return c.getTasksUC
}

// DeleteTaskUC returns DeleteTask use-case.
func (c *Container) DeleteTaskUC() commands.DeleteTaskHandler {
	if c.deleteTaskUC == nil {
		c.deleteTaskUC = services.NewDeleteTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
		)
	}

	return c.deleteTaskUC
}

// UpdateTaskUC returns UpdateTask use-case.
func (c *Container) UpdateTaskUC() commands.UpdateTaskHandler {
	if c.updateTaskUC == nil {
		c.updateTaskUC = services.NewUpdateTask(
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
