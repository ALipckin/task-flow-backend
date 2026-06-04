package app

import (
	"tasks/internal/application/ports/in/commands"
	ingrpc "tasks/internal/application/ports/in/grpc"
	"tasks/internal/application/ports/in/queries"
	"tasks/internal/application/ports/out"
	"tasks/internal/application/use_cases"
	"tasks/internal/config"
	"tasks/internal/infrastructure/adapters"
	"tasks/internal/infrastructure/sharding/shard"
	transportgrpc "tasks/internal/transport/grpc"
	grpcmiddleware "tasks/internal/transport/grpc/middleware"
	"tasks/logger"

	"google.golang.org/grpc"
)

// Container provides lazy dependency initialization for the tasks service.
type Container struct {
	cfg   *config.Config
	log   *logger.Logger
	infra *adapters.Infrastructure

	repo       out.Repository
	cache      out.Cache
	shardIndex out.TaskShardIndex
	producer   out.EventProducer
	allocator  out.IDAllocator

	createTaskUC *use_cases.CreateTask
	getTaskUC    *use_cases.GetTask
	getTasksUC   *use_cases.GetTasks
	deleteTaskUC *use_cases.DeleteTask
	updateTaskUC *use_cases.UpdateTask

	taskService ingrpc.TaskService
	grpcServer  *grpc.Server
}

// NewContainer creates a container with explicit config and logger.
func NewContainer(cfg *config.Config, log *logger.Logger) *Container {
	return &Container{cfg: cfg, log: log}
}

// Config returns the application configuration.
func (c *Container) Config() *config.Config {
	return c.cfg
}

// Logger returns the application logger.
func (c *Container) Logger() *logger.Logger {
	return c.log
}

func (c *Container) infrastructure() *adapters.Infrastructure {
	if c.infra == nil {
		c.infra = adapters.NewInfrastructure(c.cfg)
	}
	return c.infra
}

// Infrastructure returns the shared shard manager.
func (c *Container) Infrastructure() *shard.ShardManager {
	return c.infrastructure().ShardManager
}

// Repository returns the task repository implementation.
func (c *Container) Repository() out.Repository {
	if c.repo == nil {
		sm := c.Infrastructure()
		c.repo = adapters.NewPostgresRepository(
			sm,
			adapters.NewShardRouter(sm),
			c.TaskShardIndex(),
			c.Cache(),
			c.log,
		)
	}

	return c.repo
}

// TaskShardIndex returns the task shard index implementation.
func (c *Container) TaskShardIndex() out.TaskShardIndex {
	if c.shardIndex == nil {
		c.shardIndex = adapters.NewRedisTaskShardIndexAdapter(c.infrastructure().Redis)
	}

	return c.shardIndex
}

// Cache returns the cache adapter implementation.
func (c *Container) Cache() out.Cache {
	if c.cache == nil {
		c.cache = adapters.NewRedisCacheAdapter(c.infrastructure().Redis)
	}

	return c.cache
}

// Producer returns the event producer implementation.
func (c *Container) Producer() out.EventProducer {
	if c.producer == nil {
		c.producer = adapters.NewKafkaProducerAdapter(c.infrastructure().Kafka)
	}

	return c.producer
}

// Allocator returns the ID allocator implementation.
func (c *Container) Allocator() out.IDAllocator {
	if c.allocator == nil {
		c.allocator = adapters.NewRedisIDAllocator(c.infrastructure().Redis)
	}

	return c.allocator
}

// CreateTaskUC returns CreateTask use-case.
func (c *Container) CreateTaskUC() commands.CreateTaskHandler {
	if c.createTaskUC == nil {
		c.createTaskUC = use_cases.NewCreateTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
			c.Allocator(),
		)
	}

	return c.createTaskUC
}

// GetTaskUC returns GetTask use-case.
func (c *Container) GetTaskUC() queries.GetTaskHandler {
	if c.getTaskUC == nil {
		c.getTaskUC = use_cases.NewGetTask(
			c.Repository(),
			c.Cache(),
			c.Producer(),
			c.log,
		)
	}

	return c.getTaskUC
}

// GetTasksUC returns GetTasks use-case.
func (c *Container) GetTasksUC() queries.GetTasksHandler {
	if c.getTasksUC == nil {
		c.getTasksUC = use_cases.NewGetTasks(c.Repository())
	}

	return c.getTasksUC
}

// DeleteTaskUC returns DeleteTask use-case.
func (c *Container) DeleteTaskUC() commands.DeleteTaskHandler {
	if c.deleteTaskUC == nil {
		c.deleteTaskUC = use_cases.NewDeleteTask(
			c.Repository(),
			c.Cache(),
			c.TaskShardIndex(),
			c.Producer(),
		)
	}

	return c.deleteTaskUC
}

// UpdateTaskUC returns UpdateTask use-case.
func (c *Container) UpdateTaskUC() commands.UpdateTaskHandler {
	if c.updateTaskUC == nil {
		c.updateTaskUC = use_cases.NewUpdateTask(
			c.Repository(),
			c.Producer(),
		)
	}

	return c.updateTaskUC
}

// TaskService returns the inbound gRPC port with injected use-cases.
func (c *Container) TaskService() ingrpc.TaskService {
	if c.taskService == nil {
		c.taskService = &transportgrpc.TaskServer{
			CreateUC:   c.CreateTaskUC(),
			GetTaskUC:  c.GetTaskUC(),
			GetTasksUC: c.GetTasksUC(),
			DeleteUC:   c.DeleteTaskUC(),
			UpdateUC:   c.UpdateTaskUC(),
		}
	}

	return c.taskService
}

// GRPCServer returns the configured gRPC server instance.
func (c *Container) GRPCServer() *grpc.Server {
	if c.grpcServer == nil {
		c.grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpcmiddleware.UnaryLoggingInterceptor(c.log)),
		)
	}

	return c.grpcServer
}

// Cleanup releases infrastructure resources.
func (c *Container) Cleanup() {
	if c.infra != nil {
		c.infra.Close()
		c.infra = nil
	}
}
