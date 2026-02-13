FROM golang:1.22.2

# Install system dependencies (Debian-based image)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    bash \
    git \
    build-essential \
    curl \
    unzip \
    pkg-config \
    protobuf-compiler \
    libprotobuf-dev \
    libprotoc-dev \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local
# protoc is provided by package 'protobuf-compiler'

# Install Go binaries to /usr/local/bin so they're available in PATH at runtime
ENV GOBIN=/usr/local/bin

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3

ENV PATH="/root/go/bin:/usr/local/bin:$PATH"

# install CompileDaemon for live-reload during development
RUN go install github.com/githubnemo/CompileDaemon@latest

WORKDIR /app

# Copy source to image so protobufs can be generated at build time
COPY . /app

# Generate protobuf Go bindings during image build
# Add system include paths so well-known types (google/protobuf/timestamp.proto) are resolved
RUN mkdir -p ./taskpb && \
    protoc \
    --proto_path=./proto \
    --proto_path=/usr/include \
    --proto_path=/usr/local/include \
    --go_out=./taskpb \
    --go-grpc_out=./taskpb \
    proto/task.proto

# In docker-compose we mount project root into /app for live dev, so generated files may be overridden by mount.
# Run CompileDaemon to watch files and (re)build and run the package under ./cmd/tasks.
# Use explicit build flag so CompileDaemon doesn't default to building the top-level /app (which has no top-level Go files).
CMD ["CompileDaemon", "-build=go build -o /app/tasks ./cmd/tasks", "-command=/app/tasks"]
