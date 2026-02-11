FROM golang:1.24.0-alpine

RUN apk add --no-cache libc6-compat bash git

WORKDIR /app

RUN go install github.com/githubnemo/CompileDaemon@latest

CMD ["CompileDaemon", "-command=go run ./cmd/notification"]
