.PHONY: help build test lint run docker-up docker-down docker-restart docker-logs clean proto dev shell exec

help:
	@echo "TaskStorageService - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Start dev container and enter shell"
	@echo "  make shell         - Enter running container shell"
	@echo "  make exec CMD=     - Execute command in container"
	@echo "  make build         - Build the service binary (in container)"
	@echo "  make run           - Run service locally"
	@echo "  make test          - Run tests (in container)"
	@echo "  make lint          - Run linter (in container)"
	@echo "  make proto         - Regenerate protobuf files (in container)"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up     - Start service in Docker"
	@echo "  make docker-down   - Stop service in Docker"
	@echo "  make docker-restart - Restart service in Docker"
	@echo "  make docker-logs   - View service logs"
	@echo "  make docker-clean  - Stop and remove volumes"
	@echo ""
	@echo "Database:"
	@echo "  make db-reset      - Reset all shards"

dev: docker-up
	@echo "Entering dev container..."
	@docker compose exec app /bin/sh

shell:
	@echo "Entering container shell..."
	@docker compose exec app /bin/sh

exec:
	@docker compose exec app $(CMD)

build:
	@echo "Building TaskStorageService in container..."
	@docker compose exec app go build -o bin/tasks main.go
	@echo "Build complete: bin/tasks"

run:
	@echo "Starting TaskStorageService..."
	@go run main.go

test:
	@echo "Running tests in container..."
	@docker compose exec app go test ./... -v -cover

test-coverage:
	@echo "Running tests with coverage report..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || true

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

proto:
	@echo "Regenerating protobuf files in container..."
	@docker compose exec app sh -c "rm -rf ./proto/taskpb/* && \
		protoc --proto_path=proto \
		--go_out=proto/taskpb --go_opt=paths=source_relative \
		--go-grpc_out=proto/taskpb --go-grpc_opt=paths=source_relative \
		proto/task.proto"
	@echo "Protobuf files regenerated!"

docker-up:
	@echo "Starting TaskStorageService in Docker..."
	@docker compose up -d
	@echo "TaskStorageService started!"
	@echo "Waiting for shards to be healthy..."
	@sleep 5
	@docker compose ps

docker-down:
	@echo "Stopping TaskStorageService..."
	@docker compose down
	@echo "TaskStorageService stopped!"

docker-restart: docker-down docker-up

docker-logs:
	@docker compose logs -f app

docker-logs-all:
	@docker compose logs -f

docker-clean:
	@echo "Cleaning up TaskStorageService containers and volumes..."
	@docker compose down -v
	@echo "Cleanup complete!"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

install-deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

db-reset:
	@echo "Resetting all database shards..."
	@docker compose down -v
	@docker compose up -d db-shard-1 db-shard-2 db-shard-3 redis
	@echo "Database shards reset complete!"

db-logs:
	@echo "Viewing database logs..."
	@docker compose logs -f db-shard-1 db-shard-2 db-shard-3