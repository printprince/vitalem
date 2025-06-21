# Используем официальный образ Go
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata curl

# Создаем пользователя
RUN adduser -D -s /bin/sh -u 1001 appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Включаем Go modules и отключаем GOPATH
ENV GO111MODULE=on
ENV GOPATH=
ENV GOPROXY=direct
ENV GOSUMDB=off

# Скачиваем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Создаем vendor директорию со всеми зависимостями
RUN go mod vendor

# Собираем приложение используя только vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o fileserver ./cmd/fileserver

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
