.PHONY: help network up down restart logs clean build test lint proto

help:
	@echo "TaskFlow Backend - Make Commands"
	@echo ""
	@echo "Network:"
	@echo "  make network          - Create external docker network"
	@echo ""
	@echo "Service Management:"
	@echo "  make up               - Start all services"
	@echo "  make down             - Stop all services"
	@echo "  make restart          - Restart all services"
	@echo "  make logs             - View logs from all services"
	@echo "  make clean            - Stop services and remove volumes"
	@echo ""
	@echo "Individual Services:"
	@echo "  make up-auth          - Start AuthService"
	@echo "  make up-storage       - Start TaskStorageService"
	@echo "  make up-api           - Start gateway"
	@echo "  make up-notify        - Start NotifyService"
	@echo "  make down-auth        - Stop AuthService"
	@echo "  make down-storage     - Stop TaskStorageService"
	@echo "  make down-api         - Stop gateway"
	@echo "  make down-notify      - Stop NotifyService"
	@echo "  make up-api-observability - Start gateway with ELK stack"
	@echo ""
	@echo "Development:"
	@echo "  make build            - Build all services"
	@echo "  make test             - Run tests"
	@echo "  make lint             - Run linter"
	@echo "  make proto            - Regenerate protobuf files"
	@echo ""
	@echo "Database:"
	@echo "  make db-reset         - Reset all databases"

network:
	@echo "Creating external docker network..."
	@docker network create task-network 2>/dev/null || echo "Network already exists"

up: network
	@echo "Starting all services..."
	@cd AuthService && docker-compose up -d
	@cd TaskStorageService && docker-compose up -d
	@cd gateway && docker-compose up -d
	@cd NotifyService && docker-compose up -d
	@echo "All services started!"
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@make status

down:
	@echo "Stopping all services..."
	@cd NotifyService && docker-compose down
	@cd gateway && docker-compose down
	@cd TaskStorageService && docker-compose down
	@cd AuthService && docker-compose down
	@echo "All services stopped!"

restart: down up

logs:
	@echo "Streaming logs from all services..."
	@docker-compose -f AuthService/docker-compose.yml \
		-f TaskStorageService/docker-compose.yml \
		-f gateway/docker-compose.yml \
		-f NotifyService/docker-compose.yml logs -f

clean:
	@echo "Stopping services and removing volumes..."
	@cd NotifyService && docker-compose down -v
	@cd gateway && docker-compose down -v
	@cd TaskStorageService && docker-compose down -v
	@cd AuthService && docker-compose down -v
	@echo "Cleanup complete!"

up-auth: network
	@cd AuthService && make docker-up

up-storage: network
	@cd TaskStorageService && make docker-up

up-api: network
	@cd gateway && make docker-up

up-notify: network
	@cd NotifyService && make docker-up

down-auth:
	@cd AuthService && make docker-down

down-storage:
	@cd TaskStorageService && make docker-down

down-api:
	@cd gateway && make docker-down

down-notify:
	@cd NotifyService && make docker-down

up-api-observability:
	@cd gateway && make docker-up-observability

logs-auth:
	@cd AuthService && make docker-logs

logs-storage:
	@cd TaskStorageService && make docker-logs

logs-api:
	@cd gateway && make docker-logs

logs-notify:
	@cd NotifyService && make docker-logs

status:
	@echo ""
	@echo "=== Service Status ==="
	@docker ps --filter "network=task-network" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

build:
	@echo "Building all services..."
	@cd AuthService && make build
	@cd TaskStorageService && make build
	@cd gateway && make build
	@cd NotifyService && make build
	@echo "All services built successfully!"

test:
	@echo "Running tests on all services..."
	@cd AuthService && make test
	@cd TaskStorageService && make test
	@cd gateway && make test
	@cd NotifyService && make test
	@echo "All tests completed!"

lint:
	@echo "Running linter on all services..."
	@cd AuthService && make lint
	@cd TaskStorageService && make lint
	@cd gateway && make lint
	@cd NotifyService && make lint
	@echo "Linting complete!"

fmt:
	@echo "Formatting all code..."
	@cd AuthService && make fmt
	@cd TaskStorageService && make fmt
	@cd gateway && make fmt
	@cd NotifyService && make fmt
	@echo "Code formatted!"

proto:
	@echo "Regenerating protobuf files..."
	@cd TaskStorageService && make proto
	@cd gateway && make proto
	@echo "Protobuf files regenerated!"


db-reset:
	@echo "Resetting all databases..."
	@cd AuthService && make db-reset
	@cd TaskStorageService && make db-reset
	@echo "All databases reset! Run 'make up' to recreate them."

swagger:
	@echo "Regenerating Swagger documentation..."
	@cd gateway && make swagger
	@echo "Swagger docs regenerated!"

install-deps:
	@echo "Installing dependencies for all services..."
	@cd AuthService && make install-deps
	@cd TaskStorageService && make install-deps
	@cd gateway && make install-deps
	@cd NotifyService && make install-deps
	@echo "All dependencies installed!"

update-deps:
	@echo "Updating dependencies for all services..."
	@cd AuthService && make update-deps
	@cd TaskStorageService && make update-deps
	@cd gateway && make update-deps
	@cd NotifyService && make update-deps
	@echo "All dependencies updated!"

dev: network
	@echo "Starting development environment..."
	@make up
	@sleep 5
	@echo ""
	@echo "=== Development Environment Ready ==="
	@echo ""
	@make status
	@echo ""
	@echo "Services:"
	@echo "  - Auth Service:    http://localhost:5440"
	@echo "  - REST API:        http://localhost:3000"
	@echo "  - Swagger:         http://localhost:3000/swagger/index.html"
	@echo "  - Mailhog:         http://localhost:8050"
	@echo "  - Kibana:          http://localhost:5601"
	@echo "  - Elasticsearch:   http://localhost:9200"
	@echo ""
	@echo "Commands:"
	@echo "  make logs          - View all logs"
	@echo "  make status        - Show service status"
	@echo "  make down          - Stop all services"
	@echo ""
	@echo "Individual service commands:"
	@echo "  cd <service> && make help"
