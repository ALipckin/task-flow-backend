# Task storage service
Task storage service

ORM - gorm:
https://gorm.io/docs/

Installation:

create .env from .env.example

``docker compose up -d --build``

### Database sharding in Docker

Three PostgreSQL containers (shards) are launched in `docker-compose.yml`:

- **db-shard-1** — port 5432
- **db-shard-2** — port 5433
- **db-shard-3** — port 5434

Each shard has its own volume for data. The `task-storage-service` connects to all shards via the `DB_SHARD_URLS` variable (set in `environment` in docker-compose). To add or remove a shard: add/remove the `db-shard-N` service and update `DB_SHARD_URLS` in `environment` for `task-storage-service`.

to regenerate grpc taskpb files run:

``
rm -rf ./proto/taskpb/*

protoc --proto_path=proto --go_out=proto/taskpb --go_opt=paths=source_relative --go-grpc_out=proto/taskpb --go-grpc_opt=paths=source_relative proto/task.proto
``