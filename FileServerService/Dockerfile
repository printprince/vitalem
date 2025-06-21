FROM golang:1.24-alpine AS builder

WORKDIR /build

# Копируем все файлы FileServerService
COPY FileServerService/ .

# Проверяем, что файлы скопировались
RUN ls -la
RUN ls -la internal/

# Проверяем go.mod
RUN cat go.mod

# Принудительно отключаем все пути Go и используем только модули
ENV GO111MODULE=on
ENV GOPATH=""
ENV GOPROXY=direct
ENV GOSUMDB=off

# Проверяем, что Go видит модуль
RUN go list -m

# Создаем vendor директорию с зависимостями
RUN go mod download
RUN go mod vendor

# Пробуем собрать с vendor
RUN go build -mod=vendor -o fileserver ./cmd/fileserver

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