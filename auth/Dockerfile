FROM golang:1.24.0-alpine

ENV CGO_ENABLED=1

RUN apk add --no-cache \
    bash \
    git \
    gcc \
    musl-dev \
    libc6-compat \
    sqlite-dev

WORKDIR /usr/local/bin/

RUN go install github.com/githubnemo/CompileDaemon@latest

WORKDIR /usr/local/bin

CMD ["CompileDaemon", "-command=go run main.go"]
