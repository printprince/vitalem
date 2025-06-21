FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты включая git
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Копируем файлы по отдельности
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY migrations/ ./migrations/

# Полностью переопределяем Go окружение
ENV GO111MODULE=on
ENV GOPATH=""
ENV GOROOT=""
ENV GOPROXY=direct
ENV GOSUMDB=off

# Отладка Go окружения и файловой структуры
RUN go env | grep -E "(GOPATH|GOROOT|GO111MODULE|GOMOD)"
RUN pwd && ls -la
RUN echo "=== Checking internal directory ==="
RUN ls -la internal/
RUN echo "=== Checking config directory ==="
RUN ls -la internal/config/ || echo "Config directory does not exist"
RUN echo "=== Checking middleware directory ==="
RUN ls -la internal/http/middleware/ || echo "Middleware directory does not exist"
RUN echo "=== Finding all .go files ==="
RUN find . -name "*.go" | head -10

# Собираем приложение напрямую без replace директив
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