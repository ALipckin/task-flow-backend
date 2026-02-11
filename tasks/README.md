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

Each shard has its own volume for data. The `tasks-app-dev` connects to all shards via the `DB_SHARD_URLS` variable (set in `environment` in docker-compose). To add or remove a shard: add/remove the `db-shard-N` service and update `DB_SHARD_URLS` in `environment` for `tasks-app-dev`.

### Consistent Hashing and Rebalancing

- **Sharding by performer_id:** The shard for each task is determined by its `performer_id` using a consistent hash ring. The application routes requests to the appropriate PostgreSQL shard.
- **Task ID to Shard Mapping:** Stored in Redis (`task:shard:{id}`) and used for GetTask/Update/Delete operations.
- **Global Task ID:** Generated via Redis INCR counter (`task:id_counter`).
- **Adding a New Shard:** 
  1. Update the ring (restart with new `DB_SHARD_URLS` configuration)
  2. Run background rebalancing:
     - Package `rebalance`: `rebalance.Run(ctx)` performs a single pass; `rebalance.RunBackground(ctx, interval)` runs periodically in the background
     - For each `performer_id` whose shard has changed according to the ring, tasks are copied to the new shard, deleted from the old shard, and the Redis mapping is updated
- **Note:** PostgreSQL shards are unaware of sharding logic and don't perform rebalancing. All sharding logic is handled at the application layer.

to regenerate grpc taskpb files run:

``
rm -rf ./proto/taskpb/*

protoc --proto_path=proto --go_out=proto/taskpb --go_opt=paths=source_relative --go-grpc_out=proto/taskpb --go-grpc_opt=paths=source_relative proto/task.proto
``