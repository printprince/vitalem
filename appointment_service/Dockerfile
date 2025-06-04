FROM golang:1.24-alpine AS builder

WORKDIR /go/src/github.com/printprince/vitalem

# Копируем все необходимые файлы
COPY . .

# Устанавливаем GOPATH и переходим в директорию сервиса
ENV GOPATH=/go
WORKDIR /go/src/github.com/printprince/vitalem/appointment_service

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appointment_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

# Копируем исполняемый файл и конфигурацию
COPY --from=builder /go/src/github.com/printprince/vitalem/appointment_service/appointment_service .
COPY --from=builder /go/src/github.com/printprince/vitalem/appointment_service/config.yaml .

EXPOSE 8805

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8805/health || exit 1

CMD ["./appointment_service"] 