FROM golang:1.22-alpine AS builder

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
RUN cat go.mod

# Собираем приложение с принудительным использованием модулей
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