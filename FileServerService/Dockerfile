FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты включая git
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Копируем все файлы (теперь context = FileServerService)
COPY . .

# Принудительно отключаем все пути Go и используем только модули
ENV GO111MODULE=on
ENV GOPATH=""

# Собираем приложение
RUN go mod download
RUN go build -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget curl
WORKDIR /app
COPY --from=builder /app/fileserver .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8087
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1
CMD ["./fileserver"] 