# Logger Service Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем utils (общие компоненты)
COPY ./utils ./utils

# Копируем файлы logger_service
COPY ./logger_service/go.mod ./logger_service/go.sum ./
RUN go mod download

COPY ./logger_service .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o logger_service ./cmd/api

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/logger_service .
COPY ./logger_service/config.yaml .

EXPOSE 8802
CMD ["./logger_service"]