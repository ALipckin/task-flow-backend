.PHONY: help build test lint run docker-up docker-down docker-restart docker-logs clean swagger proto dev shell exec

help:
	@echo "gateway - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Start dev container and enter shell"
	@echo "  make shell         - Enter running container shell"
	@echo "  make exec CMD=     - Execute command in container"
	@echo "  make build         - Build the service binary (in container)"
	@echo "  make run           - Run service locally"
	@echo "  make test          - Run tests (in container)"
	@echo "  make lint          - Run linter (in container)"
	@echo "  make swagger       - Regenerate Swagger docs (in container)"
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
	@echo "Observability (optional ELK stack):"
	@echo "  make docker-up-observability  - Start with Elasticsearch, Kibana, Fluentd"
	@echo "  make docker-down-observability - Stop observability stack"
	@echo ""
	@echo "Before first docker-up (if container has no network):"
	@echo "  make vendor       - Create vendor/ for offline builds"

dev: docker-up
	@echo "Entering dev container..."
	@docker compose exec gateway-app /bin/sh

shell:
	@echo "Entering container shell..."
	@docker compose exec gateway-app /bin/sh

exec:
	@docker compose exec gateway-app $(CMD)

build:
	@echo "Building gateway in container..."
	@docker compose exec gateway-app go build -o bin/rest-api main.go
	@echo "Build complete: bin/rest-api"

run:
	@echo "Starting gateway..."
	@go run main.go

test:
	@echo "Running tests in container..."
	@docker compose exec gateway-app go test ./... -v -cover

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || true

fmt:
	@go fmt ./...

swagger:
	@echo "Regenerating Swagger documentation in container..."
	@docker compose exec gateway-app swag init
	@echo "Swagger docs regenerated!"

proto:
	@echo "Regenerating protobuf files in container..."
	@docker compose exec gateway-app sh -c "rm -rf ./proto/taskpb/* && protoc --proto_path=proto --go_out=proto/taskpb --go_opt=paths=source_relative --go-grpc_out=proto/taskpb --go-grpc_opt=paths=source_relative proto/task.proto"
	@echo "Protobuf files regenerated!"

docker-up:
	@echo "Starting gateway in Docker..."
	@docker compose up -d
	@echo "gateway started!"
	@echo ""
	@echo "Services: REST API and Swagger on PORT (default 3000)"
	@echo "For ELK stack: make docker-up-observability"

docker-down:
	@echo "Stopping gateway..."
	@docker compose down
	@echo "gateway stopped!"

docker-restart: docker-down docker-up

docker-up-observability:
	@echo "Starting gateway with ELK stack..."
	@docker compose -f docker-compose.yml -f docker-compose.observability.yml up -d
	@echo "Started! Elasticsearch: 9200, Kibana: 5601, Fluentd: 24224"

docker-down-observability:
	@echo "Stopping observability stack (Elasticsearch, Kibana, Fluentd)..."
	@docker compose -f docker-compose.observability.yml down

docker-logs:
	@docker compose logs -f app

docker-logs-all:
	@docker compose logs -f

docker-clean:
	@docker compose down -v

clean:
	@rm -rf bin/

vendor:
	@echo "Vendoring dependencies..."
	@go mod vendor
	@echo "Run: make docker-up"

install-deps:
	@go mod download
	@go mod tidy

kafka-topics:
	@docker compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-consume:
	@docker compose exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic task_events --from-beginning
