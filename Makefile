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
	@echo "  make up-auth          - Start auth"
	@echo "  make up-storage       - Start tasks"
	@echo "  make up-api           - Start gateway"
	@echo "  make up-notify        - Start notification"
	@echo "  make down-auth        - Stop auth"
	@echo "  make down-storage     - Stop tasks"
	@echo "  make down-api         - Stop gateway"
	@echo "  make down-notify      - Stop notification"
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
	@cd auth && docker-compose up -d
	@cd tasks && docker-compose up -d
	@cd gateway && docker-compose up -d
	@cd notification && docker-compose up -d
	@echo "All services started!"
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@make status

down:
	@echo "Stopping all services..."
	@cd notification && docker-compose down
	@cd gateway && docker-compose down
	@cd tasks && docker-compose down
	@cd auth && docker-compose down
	@echo "All services stopped!"

restart: down up

logs:
	@echo "Streaming logs from all services..."
	@docker-compose -f auth/docker-compose.yml \
		-f tasks/docker-compose.yml \
		-f gateway/docker-compose.yml \
		-f notification/docker-compose.yml logs -f

clean:
	@echo "Stopping services and removing volumes..."
	@cd notification && docker-compose down -v
	@cd gateway && docker-compose down -v
	@cd tasks && docker-compose down -v
	@cd auth && docker-compose down -v
	@echo "Cleanup complete!"

up-auth: network
	@cd auth && make docker-up

up-storage: network
	@cd tasks && make docker-up

up-api: network
	@cd gateway && make docker-up

up-notify: network
	@cd notification && make docker-up

down-auth:
	@cd auth && make docker-down

down-storage:
	@cd tasks && make docker-down

down-api:
	@cd gateway && make docker-down

down-notify:
	@cd notification && make docker-down

up-api-observability:
	@cd gateway && make docker-up-observability

logs-auth:
	@cd auth && make docker-logs

logs-storage:
	@cd tasks && make docker-logs

logs-api:
	@cd gateway && make docker-logs

logs-notify:
	@cd notification && make docker-logs

status:
	@echo ""
	@echo "=== Service Status ==="
	@docker ps --filter "network=task-network" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

build:
	@echo "Building all services..."
	@cd auth && make build
	@cd tasks && make build
	@cd gateway && make build
	@cd notification && make build
	@echo "All services built successfully!"

test:
	@echo "Running tests on all services..."
	@cd auth && make test
	@cd tasks && make test
	@cd gateway && make test
	@cd notification && make test
	@echo "All tests completed!"

lint:
	@echo "Running linter on all services..."
	@cd auth && make lint
	@cd tasks && make lint
	@cd gateway && make lint
	@cd notification && make lint
	@echo "Linting complete!"

fmt:
	@echo "Formatting all code..."
	@cd auth && make fmt
	@cd tasks && make fmt
	@cd gateway && make fmt
	@cd notification && make fmt
	@echo "Code formatted!"

proto:
	@echo "Regenerating protobuf files..."
	@cd tasks && make proto
	@cd gateway && make proto
	@echo "Protobuf files regenerated!"


db-reset:
	@echo "Resetting all databases..."
	@cd auth && make db-reset
	@cd tasks && make db-reset
	@echo "All databases reset! Run 'make up' to recreate them."

swagger:
	@echo "Regenerating Swagger documentation..."
	@cd gateway && make swagger
	@echo "Swagger docs regenerated!"

install-deps:
	@echo "Installing dependencies for all services..."
	@cd auth && make install-deps
	@cd tasks && make install-deps
	@cd gateway && make install-deps
	@cd notification && make install-deps
	@echo "All dependencies installed!"

update-deps:
	@echo "Updating dependencies for all services..."
	@cd auth && make update-deps
	@cd tasks && make update-deps
	@cd gateway && make update-deps
	@cd notification && make update-deps
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
