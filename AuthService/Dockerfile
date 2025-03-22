FROM golang:1.24.0-alpine

RUN apk add --no-cache libc6-compat bash git

WORKDIR /usr/local/bin/

RUN go install github.com/githubnemo/CompileDaemon@latest

WORKDIR /usr/local/bin

CMD ["CompileDaemon", "-command=go run main.go"]
