FROM golang:1.24-alpine AS builder

WORKDIR /build

# Копируем все файлы FileServerService
COPY FileServerService/ .

# Проверяем, что файлы скопировались
RUN ls -la
RUN ls -la internal/

# Проверяем go.mod
RUN cat go.mod

# Создаем go.work для принудительной работы с модулями
RUN echo "go 1.24" > go.work
RUN echo "" >> go.work
RUN echo "use ." >> go.work

# Проверяем, что Go видит модуль
RUN go list -m
RUN GOWORK=go.work go build -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget curl
WORKDIR /app
COPY --from=builder /build/fileserver .
COPY --from=builder /build/migrations ./migrations
EXPOSE 8087
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1
CMD ["./fileserver"] 