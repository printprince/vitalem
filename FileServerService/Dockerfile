# Используем официальный образ Go
FROM golang:1.24-alpine AS builder

WORKDIR /go/src/github.com/printprince/vitalem

# Копируем все необходимые файлы
COPY . .

# Устанавливаем GOPATH и переходим в директорию сервиса
ENV GOPATH=/go
WORKDIR /go/src/github.com/printprince/vitalem/FileServerService

# Отключаем Go modules для использования GOPATH
ENV GO111MODULE=off

# Устанавливаем зависимости через go get
RUN go get -d ./...

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Копируем исполняемый файл
COPY --from=builder /go/src/github.com/printprince/vitalem/FileServerService/fileserver .

# Копируем миграции
COPY --from=builder /go/src/github.com/printprince/vitalem/FileServerService/migrations ./migrations

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./fileserver"]
