FROM golang:1.22.2-alpine

RUN apk add --no-cache \
    libc6-compat \
    bash \
    git \
    build-base \
    protobuf-dev \
    curl \
    unzip

WORKDIR /usr/local
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v21.8/protoc-21.8-linux-x86_64.zip \
    && unzip protoc-21.8-linux-x86_64.zip \
    && rm protoc-21.8-linux-x86_64.zip

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3

ENV PATH="/root/go/bin:$PATH"
WORKDIR /usr/local/bin/

RUN go install github.com/githubnemo/CompileDaemon@latest

WORKDIR /usr/local/bin

CMD ["CompileDaemon", "-command=go run main.go"]