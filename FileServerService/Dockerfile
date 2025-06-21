# Используем официальный образ Go
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Копируем все файлы
COPY . .

# Простая сборка с модулями
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates curl

# Создаем пользователя
RUN adduser -D -s /bin/sh -u 1001 appuser

# Создаем директории
RUN mkdir -p /app && chown appuser:appuser /app

# Копируем бинарник
COPY --from=builder --chown=appuser:appuser /app/fileserver /app/fileserver

# Копируем миграции
COPY --from=builder --chown=appuser:appuser /app/migrations /app/migrations

# Переключаемся на пользователя
USER appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Запускаем приложение
CMD ["./fileserver"]
