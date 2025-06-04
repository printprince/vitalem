FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY appointment_service/go.mod appointment_service/go.sum ./
RUN go mod download

# Копируем исходный код
COPY appointment_service/ .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appointment_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

# Копируем исполняемый файл и конфигурацию
COPY --from=builder /app/appointment_service .
COPY --from=builder /app/config.yaml .

EXPOSE 8805

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8805/health || exit 1

CMD ["./appointment_service"] 