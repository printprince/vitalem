FROM golang:1.24-alpine AS builder

WORKDIR /go/src/github.com/printprince/vitalem

# Копируем все необходимые файлы
COPY . .

# Принудительно включаем Go modules и полностью отключаем GOPATH
ENV GO111MODULE=on
ENV GOPATH=""

# Обновляем go.mod и go.sum
RUN cd FileServerService && go mod tidy && go mod download

# Собираем приложение с Go modules
RUN cd FileServerService && CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Копируем исполняемый файл и миграции
COPY --from=builder /go/src/github.com/printprince/vitalem/FileServerService/fileserver .
COPY --from=builder /go/src/github.com/printprince/vitalem/FileServerService/migrations ./migrations

EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1

# Запускаем приложение
CMD ["./fileserver"] 