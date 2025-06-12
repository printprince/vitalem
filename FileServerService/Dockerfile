# Build stage
FROM golang:1.21-alpine AS builder

# Установка git и необходимых инструментов
RUN apk add --no-cache git gcc musl-dev

# Установка рабочей директории
WORKDIR /app

# Копирование файлов зависимостей
COPY go.mod ./
COPY go.sum ./

# Загрузка зависимостей
RUN go mod download && go mod verify

# Копирование исходного кода
COPY . ./

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/fileserver

# Final stage
FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование бинарного файла из builder
COPY --from=builder /app/fileserver ./

# Открытие порта
EXPOSE 8081

# Запуск приложения
CMD ["./fileserver"]
