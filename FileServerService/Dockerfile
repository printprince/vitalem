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

# Создаем go.mod файлы для каждого internal пакета
RUN echo 'module fileserver/internal/config' > internal/config/go.mod
RUN echo 'go 1.24' >> internal/config/go.mod

RUN echo 'module fileserver/internal/http/middleware' > internal/http/middleware/go.mod
RUN echo 'go 1.24' >> internal/http/middleware/go.mod

RUN echo 'module fileserver/internal/http/logger' > internal/http/logger/go.mod
RUN echo 'go 1.24' >> internal/http/logger/go.mod

RUN echo 'module fileserver/internal/http/router' > internal/http/router/go.mod
RUN echo 'go 1.24' >> internal/http/router/go.mod

RUN echo 'module fileserver/internal/repository' > internal/repository/go.mod
RUN echo 'go 1.24' >> internal/repository/go.mod

RUN echo 'module fileserver/internal/service' > internal/service/go.mod
RUN echo 'go 1.24' >> internal/service/go.mod

RUN echo 'module fileserver/internal/storage' > internal/storage/go.mod
RUN echo 'go 1.24' >> internal/storage/go.mod

RUN echo 'module fileserver/internal/model' > internal/model/go.mod
RUN echo 'go 1.24' >> internal/model/go.mod

RUN echo 'module fileserver/internal/http/handler' > internal/http/handler/go.mod
RUN echo 'go 1.24' >> internal/http/handler/go.mod

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