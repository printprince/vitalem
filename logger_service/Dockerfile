# Logger Service Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Создаем структуру директорий для правильной работы с replace директивами
COPY ./utils ./utils
COPY ./logger_service ./logger_service

# Переходим в директорию сервиса для сборки
WORKDIR /app/logger_service

RUN go mod download && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o logger_service ./cmd/api

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/logger_service/logger_service .
COPY --from=builder /app/logger_service/config.yaml .

EXPOSE 8802
CMD ["./logger_service"]