# Используем официальный образ Go 1.22 как базовый
FROM golang:1.22.2-alpine

# Устанавливаем необходимые библиотеки для работы с Go бинарниками
RUN apk add --no-cache libc6-compat bash git

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /usr/local/bin/

# Устанавливаем CompileDaemon для автоматической компиляции и запуска приложения
RUN go install github.com/githubnemo/CompileDaemon@latest

# Устанавливаем рабочую директорию
WORKDIR /usr/local/bin

# Указываем команду для запуска приложения
CMD ["CompileDaemon", "-command=go run main.go"]
