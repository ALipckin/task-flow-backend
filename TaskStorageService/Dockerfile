# Используем официальный образ Go 1.22 как базовый
FROM golang:1.22.2-alpine

# Устанавливаем зависимости
RUN apk add --no-cache \
    libc6-compat \
    bash \
    git \
    build-base \
    protobuf-dev \
    curl \
    unzip

# Устанавливаем protoc (Protocol Buffers)
WORKDIR /usr/local
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v21.8/protoc-21.8-linux-x86_64.zip \
    && unzip protoc-21.8-linux-x86_64.zip \
    && rm protoc-21.8-linux-x86_64.zip

# Устанавливаем плагины для Go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3

ENV PATH="/root/go/bin:$PATH"
# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /usr/local/bin/

# Устанавливаем CompileDaemon для автоматической компиляции и запуска приложения
RUN go install github.com/githubnemo/CompileDaemon@latest

# Устанавливаем рабочую директорию
WORKDIR /usr/local/bin

# Указываем команду для запуска приложения
CMD ["CompileDaemon", "-command=go run main.go"]
