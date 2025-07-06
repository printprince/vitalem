# Используем официальный Golang образ для сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем весь проект (нужно для локальных зависимостей)
COPY . .

# Переходим в директорию gateway_service
WORKDIR /app/gateway_service

# Загружаем зависимости
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/api

# Финальная стадия - используем минимальный образ
FROM alpine:latest

# Добавляем ca-certificates для HTTPS запросов и wget для health check
RUN apk --no-cache add ca-certificates wget

# Создаем пользователя для безопасности
RUN addgroup -g 1000 -S gateway && \
    adduser -u 1000 -S gateway -G gateway

WORKDIR /app

# Копируем бинарник, конфигурацию и Swagger документацию
COPY --from=builder /app/gateway_service/gateway .
COPY --from=builder /app/gateway_service/config.yaml .
COPY --from=builder /app/gateway_service/docs ./docs

# Меняем владельца файлов
RUN chown -R gateway:gateway /app

# Переключаемся на пользователя gateway
USER gateway

# Открываем порт
EXPOSE 8800

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8800/health || exit 1

# Запускаем приложение
CMD ["./gateway"] 