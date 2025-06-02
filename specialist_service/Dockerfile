FROM golang:1.24-alpine AS builder

WORKDIR /app

# Создаем структуру директорий для правильной работы с replace директивами
COPY ./utils ./utils
COPY ./specialist_service ./specialist_service

# Переходим в директорию сервиса для сборки
WORKDIR /app/specialist_service

RUN go mod download && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o specialist_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

COPY --from=builder /app/specialist_service/specialist_service .
COPY --from=builder /app/specialist_service/config.yaml .

EXPOSE 8803

CMD ["./specialist_service"]
