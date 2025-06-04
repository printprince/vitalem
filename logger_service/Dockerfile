# Logger Service Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY logger_service/go.mod logger_service/go.sum ./
RUN go mod download

# Копируем исходный код
COPY logger_service/ .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o logger_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Копируем исполняемый файл и конфигурацию
COPY --from=builder /app/logger_service .
COPY --from=builder /app/config.yaml .

EXPOSE 8802

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8802/health || exit 1

CMD ["./logger_service"]