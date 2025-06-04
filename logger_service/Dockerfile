# Logger Service Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /go/src/github.com/printprince/vitalem

# Копируем все необходимые файлы
COPY . .

# Устанавливаем GOPATH и переходим в директорию сервиса
ENV GOPATH=/go
WORKDIR /go/src/github.com/printprince/vitalem/logger_service

# Обновляем go.mod и go.sum
RUN go mod tidy
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o logger_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Копируем исполняемый файл и конфигурацию
COPY --from=builder /go/src/github.com/printprince/vitalem/logger_service/logger_service .
COPY --from=builder /go/src/github.com/printprince/vitalem/logger_service/config.yaml .

EXPOSE 8802

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8802/health || exit 1

CMD ["./logger_service"]