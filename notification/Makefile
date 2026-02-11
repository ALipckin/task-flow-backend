.PHONY: help build test lint run docker-up docker-down docker-restart docker-logs clean dev shell exec

help:
	@echo "NotifyService - Available Commands"
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
	@echo ""
	@echo "Email:"
	@echo "  make mailhog       - Open Mailhog UI"

dev: docker-up
	@echo "Entering dev container..."
	@docker compose exec app /bin/sh

shell:
	@echo "Entering container shell..."
	@docker compose exec app /bin/sh

exec:
	@docker compose exec app $(CMD)

build:
	@echo "Building NotifyService in container..."
	@docker compose exec app go build -o bin/notify ./cmd/notification
	@echo "Build complete: bin/notify"

run:
	@echo "Starting NotifyService..."
	@go run ./cmd/notification

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

docker-up:
	@echo "Starting NotifyService in Docker..."
	@docker compose up -d
	@echo "NotifyService started!"
	@echo ""
	@echo "Mailhog UI: http://localhost:8050 (or check MAILHOG_UI_PORT in .env)"

docker-down:
	@echo "Stopping NotifyService..."
	@docker compose down
	@echo "NotifyService stopped!"

docker-restart: docker-down docker-up

docker-logs:
	@docker compose logs -f app

docker-logs-all:
	@docker compose logs -f

docker-clean:
	@echo "Cleaning up NotifyService containers and volumes..."
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

mailhog:
	@echo "Opening Mailhog UI..."
	@open http://localhost:8050 || xdg-open http://localhost:8050 2>/dev/null || echo "Please open http://localhost:8050 in your browser"
