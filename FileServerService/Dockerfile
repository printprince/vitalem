FROM golang:1.22-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Принудительно включаем Go modules
ENV GO111MODULE=on

# Копируем только FileServerService
COPY FileServerService/ .

# Переинициализируем модуль
RUN rm -f go.mod go.sum
RUN go mod init fileserver
RUN go mod tidy

# Отладочная информация
RUN pwd
RUN ls -la
RUN ls -la internal/
RUN find internal -type d
RUN find internal -name "*.go" | head -10

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Копируем исполняемый файл и миграции
COPY --from=builder /app/fileserver .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1

# Запускаем приложение
CMD ["./fileserver"] 