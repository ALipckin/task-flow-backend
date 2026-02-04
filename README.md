# TaskFlow Backend

A microservices-based task management system built with Go, featuring database sharding, event-driven architecture, and real-time notifications.

## Architecture Overview

TaskFlow Backend is built using a microservices architecture with the following services:

```
┌─────────────────┐
│   REST API      │ ← Gin Framework (HTTP/WebSocket)
│   Gateway       │
└────────┬────────┘
         │
    ┌────┴────┬─────────────┬──────────────┐
    │         │             │              │
┌───▼───┐ ┌──▼──────┐ ┌────▼─────┐ ┌─────▼──────┐
│ Auth  │ │  Task   │ │  Notify  │ │   Kafka    │
│Service│ │ Storage │ │ Service  │ │  Message   │
│       │ │ Service │ │          │ │   Broker   │
└───┬───┘ └────┬────┘ └────┬─────┘ └────────────┘
    │          │           │
    │     ┌────┴───────────┴───┐
    │     │                    │
┌───▼─────▼──┐          ┌─────▼──────┐
│ PostgreSQL │          │   Redis    │
│ (Auth DB + │          │  (Cache +  │
│ 3 Shards)  │          │  Task IDs) │
└────────────┘          └────────────┘
```

## Services

### 1. TaskRestApiService (Gateway)
- **Technology**: Gin (Go HTTP framework)
- **Port**: 3000 (configurable)
- **Responsibilities**:
  - REST API gateway for all client requests
  - Authentication middleware
  - Rate limiting
  - WebSocket connections for real-time notifications
  - Swagger documentation
  - Request proxying to backend services

### 2. TaskStorageService (Core)
- **Technology**: gRPC
- **Port**: 50051
- **Responsibilities**:
  - Task CRUD operations
  - Database sharding by `performer_id` using consistent hashing
  - Task ID generation via Redis
  - Cache management
  - Event publishing to Kafka
  - Automatic rebalancing on shard addition

**Key Features**:
- **Consistent Hashing**: Tasks distributed across 3 PostgreSQL shards based on `performer_id`
- **Global Task IDs**: Generated via Redis INCR counter
- **Task-to-Shard Mapping**: Stored in Redis for quick lookup
- **Dynamic Rebalancing**: Background process for redistributing tasks when shards are added

### 3. AuthService
- **Technology**: Standard HTTP
- **Port**: 8081
- **Responsibilities**:
  - User registration and login
  - JWT token generation and validation
  - User management
  - Password hashing (bcrypt)

### 4. NotifyService
- **Technology**: Kafka Consumer
- **Responsibilities**:
  - Consuming task events from Kafka
  - Sending email notifications via SMTP
  - Processing task events (created, updated, deleted)

## Infrastructure

### Databases
- **PostgreSQL 14**:
  - Auth Database (port 5435)
  - Task Shard 1 (port 5432)
  - Task Shard 2 (port 5433)
  - Task Shard 3 (port 5434)

### Message Broker
- **Apache Kafka 7.5.3** with Zookeeper
  - Topics: `task_events`, `notify_events`
  - Port: 9092

### Cache & Storage
- **Redis**: Task caching and ID generation (port 6379)

### Monitoring & Logging
- **Elasticsearch**: Log aggregation (port 9200)
- **Kibana**: Log visualization (port 5601)
- **Fluentd**: Log collection (port 24224)
- **Mailhog**: Email testing (SMTP: 1025, UI: 8050)

## Prerequisites

- **Docker** >= 20.10
- **Docker Compose** >= 2.0
- **Go** >= 1.21 (for local development)
- **Make** (for using Makefile commands)
- **protoc** (for regenerating protobuf files)

## Quick Start

### Option 1: Quick Start (Recommended)
```bash
make dev
```

This single command will:
- Create the Docker network
- Start all services
- Wait for health checks
- Show service status
- Display all service URLs

### Option 2: Manual Start

#### 1. Create Docker Network
```bash
make network
```

#### 2. Start All Services
```bash
make up
```

This will start:
- All databases with health checks
- Kafka and Zookeeper
- Redis
- All microservices
- Monitoring stack (Elasticsearch, Kibana)

#### 3. Check Service Status
```bash
make status
```

#### 4. View Logs
```bash
# From root directory
make logs              # All services
make logs-api          # REST API only
make logs-storage      # Task Storage only
make logs-auth         # Auth Service only
make logs-notify       # Notify Service only

# From service directory
cd TaskRestApiService && make docker-logs
```

### 5. Access Services

- **REST API**: http://localhost:3000
- **Swagger Docs**: http://localhost:3000/swagger/index.html
- **Auth Service**: http://localhost:5440
- **Kibana**: http://localhost:5601
- **Elasticsearch**: http://localhost:9200
- **Mailhog UI**: http://localhost:8050

## Makefile Commands

Each service has its own Makefile for independent development. The root Makefile orchestrates all services.

### Root Makefile (All Services)

```bash
# Quick Start
make dev             # Start development environment with status

# Service Management
make up              # Start all services
make down            # Stop all services
make restart         # Restart all services
make status          # Show service status
make logs            # Stream logs from all services

# Development
make build           # Build all services
make test            # Run all tests
make lint            # Run linter on all services
make fmt             # Format all code

# Utilities
make proto           # Regenerate protobuf files
make swagger         # Regenerate Swagger docs
make db-reset        # Reset all databases
```

### Individual Service Makefiles

Every service has its own Makefile. Example commands:

```bash
# Navigate to service
cd AuthService       # or TaskStorageService, TaskRestApiService, NotifyService

# Common commands (available in all services)
make help            # Show all available commands
make build           # Build service binary
make run             # Run service locally
make test            # Run tests
make test-coverage   # Run tests with coverage report
make lint            # Run linter
make fmt             # Format code
make docker-up       # Start service in Docker
make docker-down     # Stop Docker containers
make docker-logs     # View service logs
make clean           # Clean build artifacts
```

### Service-Specific Commands

**TaskStorageService**:
```bash
cd TaskStorageService
make proto           # Regenerate protobuf files
make db-reset        # Reset all database shards
make db-logs         # View database logs
```

**TaskRestApiService**:
```bash
cd TaskRestApiService
make swagger         # Regenerate Swagger docs
make proto           # Regenerate protobuf files
make kafka-topics    # List Kafka topics
make kafka-consume   # Consume from Kafka
```

**NotifyService**:
```bash
cd NotifyService
make mailhog         # Open Mailhog UI
```

## API Documentation

### Authentication

#### Register
```bash
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "securepassword"
}
```

#### Login
```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "email": "user@example.com",
  "name": "John Doe",
  "id": "1"
}
```

### Tasks

All task endpoints require authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your-token>
```

#### Create Task
```bash
POST /tasks
Content-Type: application/json

{
  "title": "Complete project",
  "description": "Finish the backend implementation",
  "performer_id": 2,
  "creator_id": 1,
  "observer_ids": [3, 4],
  "status": "pending"
}
```

#### Get Tasks
```bash
GET /tasks?performer_id=2&status=pending
```

#### Get Task by ID
```bash
GET /tasks/1
```

#### Update Task
```bash
PUT /tasks/1
Content-Type: application/json

{
  "title": "Updated title",
  "description": "Updated description",
  "performer_id": 2,
  "creator_id": 1,
  "status": "in_progress"
}
```

#### Delete Task
```bash
DELETE /tasks/1
```

### WebSocket Notifications

Connect to WebSocket for real-time task notifications:

```javascript
const ws = new WebSocket('ws://localhost:3000/tasks/notifications');

// Authenticate
ws.send(JSON.stringify({
  type: 'authenticate',
  token: 'your-jwt-token'
}));

// Receive notifications
ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('Task event:', notification);
};
```

## Development Guide

### Local Development

#### Option 1: Run Everything in Docker (Recommended)
```bash
# Start all services with hot reload
make dev

# Or start individual service
cd TaskRestApiService
make docker-up
```

Services use `CompileDaemon` for automatic reloading on code changes.

#### Option 2: Run Service Locally (for debugging)

1. **Start Infrastructure**:
```bash
# Start dependencies for a specific service
cd TaskStorageService
make docker-up  # Starts databases, Redis, Kafka
```

2. **Run Service Locally**:
```bash
# In another terminal
cd TaskRestApiService
make run  # or: go run main.go
```

#### Working with Individual Services

Each service is independent and can be developed separately:

```bash
# Example: Working on AuthService
cd AuthService
make help           # See all available commands
make build          # Build the service
make test           # Run tests
make docker-up      # Start in Docker
make docker-logs    # View logs
```

### Adding a New Shard

1. **Update docker-compose.yml**:
```yaml
db-shard-4:
  image: postgres:14
  container_name: postgres-shard-4
  environment:
    POSTGRES_USER: user
    POSTGRES_PASSWORD: password
    POSTGRES_DB: tasks_db
  ports:
    - "5437:5432"
  volumes:
    - postgres-shard-4-data:/var/lib/postgresql/data
```

2. **Update DB_SHARD_URLS** in `.env`:
```env
DB_SHARD_URLS=...existing...,postgres://user:password@db-shard-4:5432/tasks_db
```

3. **Restart and Rebalance**:
```bash
make restart
# Rebalancing happens automatically in the background
```

### Running Tests

```bash
# All services (from root)
make test

# Specific service
cd AuthService
make test

# With coverage report
cd AuthService
make test-coverage  # Generates coverage.html
```

### Linting

```bash
# All services (from root)
make lint

# Specific service
cd TaskStorageService
make lint
```

### Formatting Code

```bash
# All services (from root)
make fmt

# Specific service
cd TaskStorageService
make fmt
```

### Regenerating Protobuf Files

```bash
# All services with proto files (from root)
make proto

# Specific service
cd TaskStorageService
make proto
```

## Debugging

### View Logs
```bash
# Real-time logs
make logs

# Last 100 lines
docker logs task-storage-service --tail 100

# Follow logs
docker logs -f task-rest-api-service
```

### Check Health
```bash
# Service status
make status

# Database health
docker exec postgres-shard-1 pg_isready -U user

# Redis health
docker exec redis redis-cli ping

# Kafka health
docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### Access Kibana
1. Open http://localhost:5601
2. Create index pattern: `logstash-*`
3. View logs in Discover tab

## Troubleshooting

### Services Won't Start

**Check dependencies**:
```bash
make status
docker ps -a
```

**Common issues**:
- Ports already in use: Change ports in `.env` files
- Network doesn't exist: Run `make network`
- Database not ready: Wait for health checks

### Build Errors

**Clear and rebuild**:
```bash
make clean
make up
```

**Check Go modules**:
```bash
cd <service>
go mod tidy
go mod download
```

### Database Connection Issues

**Reset databases**:
```bash
make db-reset
make up
```

**Check connection strings**:
```bash
docker exec postgres-shard-1 psql -U user -d tasks_db -c "\l"
```

### Kafka Issues

**Restart Kafka**:
```bash
docker-compose -f TaskRestApiService/docker-compose.yml restart kafka zookeeper
```

**Check topics**:
```bash
docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

## Security

### Best Practices Implemented
- JWT tokens with expiration
- Password hashing with bcrypt
- Rate limiting on API endpoints
- CORS configuration
- Input validation
- SQL injection protection (via GORM)

### Security Checklist
- [ ] Change default passwords in `.env` files
- [ ] Use strong JWT secrets (256+ bits)
- [ ] Enable SSL/TLS for production
- [ ] Configure proper CORS origins
- [ ] Regular dependency updates
- [ ] Security audit logs
