FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты включая git
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Копируем все файлы (теперь context = FileServerService)
COPY . .

# Полностью переопределяем Go окружение
ENV GO111MODULE=on
ENV GOPATH=""
ENV GOROOT=""
ENV GOPROXY=direct
ENV GOSUMDB=off

# Отладка Go окружения
RUN go env | grep -E "(GOPATH|GOROOT|GO111MODULE|GOMOD)"
RUN pwd && ls -la

# Создаем новый go.mod с replace директивами
RUN cp go.mod go.mod.backup
RUN echo 'module fileserver' > go.mod.new
RUN echo '' >> go.mod.new
RUN echo 'go 1.24' >> go.mod.new
RUN echo '' >> go.mod.new
RUN echo 'replace fileserver/internal/config => ./internal/config' >> go.mod.new
RUN echo 'replace fileserver/internal/http/middleware => ./internal/http/middleware' >> go.mod.new
RUN echo 'replace fileserver/internal/http/logger => ./internal/http/logger' >> go.mod.new
RUN echo 'replace fileserver/internal/http/router => ./internal/http/router' >> go.mod.new
RUN echo 'replace fileserver/internal/repository => ./internal/repository' >> go.mod.new
RUN echo 'replace fileserver/internal/service => ./internal/service' >> go.mod.new
RUN echo 'replace fileserver/internal/storage => ./internal/storage' >> go.mod.new
RUN echo 'replace fileserver/internal/model => ./internal/model' >> go.mod.new
RUN echo '' >> go.mod.new
RUN tail -n +4 go.mod.backup >> go.mod.new
RUN mv go.mod.new go.mod

# Собираем приложение с новым go.mod
RUN go mod tidy
RUN go mod download
RUN go build -mod=mod -o fileserver ./cmd/fileserver

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