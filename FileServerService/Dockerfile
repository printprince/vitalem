FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Полностью переопределяем Go окружение
ENV GO111MODULE=on
ENV GOPATH=""
ENV GOPROXY=direct
ENV GOSUMDB=off

# Копируем только FileServerService
COPY FileServerService/ .

# Скачиваем зависимости
RUN go mod download

# Собираем приложение с максимальной изоляцией от стандартных путей
RUN cd /app && CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o fileserver ./cmd/fileserver

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