.PHONY: help build test lint run docker-up docker-down docker-restart docker-logs clean dev shell exec

help:
	@echo "AuthService - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Start dev container and enter shell"
	@echo "  make shell         - Enter running container shell"
	@echo "  make exec CMD=     - Execute command in container"
	@echo "  make build         - Build the service binary (in container)"
	@echo "  make run           - Run service locally"
	@echo "  make test          - Run tests (in container)"
	@echo "  make lint          - Run linter (in container)"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up     - Start service in Docker"
	@echo "  make docker-down   - Stop service in Docker"
	@echo "  make docker-restart - Restart service in Docker"
	@echo "  make docker-logs   - View service logs"
	@echo "  make docker-clean  - Stop and remove volumes"

dev: docker-up
	@echo "Entering dev container..."
	@docker exec -it auth-service-app /bin/sh

shell:
	@echo "Entering container shell..."
	@docker exec -it auth-service-app /bin/sh

exec:
	@docker exec -it auth-service-app $(CMD)

build:
	@echo "Building AuthService in container..."
	@docker exec -it auth-service-app go build -o bin/auth main.go
	@echo "Build complete: bin/auth"

run:
	@echo "Starting AuthService..."
	@go run main.go

test:
	@echo "Running tests in container..."
	@docker exec -it auth-service-app go test ./... -v -cover

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

docker-up:
	@echo "Starting AuthService in Docker..."
	@docker-compose up -d
	@echo "AuthService started!"

docker-down:
	@echo "Stopping AuthService..."
	@docker-compose down
	@echo "AuthService stopped!"

docker-restart: docker-down docker-up

docker-logs:
	@docker-compose logs -f

docker-clean:
	@echo "Cleaning up AuthService containers and volumes..."
	@docker-compose down -v
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
	@echo "Resetting database..."
	@docker-compose down -v
	@docker-compose up -d auth-db
	@echo "Database reset complete!"
