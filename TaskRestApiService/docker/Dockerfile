FROM golang:1.24-alpine

RUN apk add --no-cache libc6-compat bash git

WORKDIR /usr/local/bin/

RUN go install github.com/githubnemo/CompileDaemon@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /usr/local/bin
