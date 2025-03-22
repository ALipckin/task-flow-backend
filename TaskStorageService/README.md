# Task storage service
Task storage service

ORM - gorm:
https://gorm.io/docs/

Installation:

create .env from .evn-example

``docker compose up -d --build``

to generate grpc taskpb files run:

``
protoc --proto_path=proto   --go_out=proto/taskpb --go_opt=paths=source_relative   --go-grpc_out=proto/taskpb --go-grpc_opt=paths=source_relative   proto/task.proto
``