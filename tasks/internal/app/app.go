package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tasks/internal/config"
	"tasks/logger"
	"tasks/proto/taskpb"
)

// App owns the application lifecycle.
type App struct {
	container *Container
}

// New creates an App from config and logger.
func New(cfg *config.Config, log *logger.Logger) *App {
	return &App{container: NewContainer(cfg, log)}
}

// Run starts the gRPC server and blocks until shutdown.
func (a *App) Run() error {
	cfg := a.container.Config()
	grpcServer := a.container.GRPCServer()
	taskpb.RegisterTaskServiceServer(grpcServer, a.container.TaskServer())

	port := ":" + cfg.GRPCPort
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("start error: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("gRPC-server start", "port", port)
		serveErr <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("gRPC server error: %w", err)
	case sig := <-sigChan:
		slog.Info("received signal, initiating graceful shutdown", "signal", sig.String())
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		slog.Info("shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	}

	a.container.Cleanup()
	return nil
}
