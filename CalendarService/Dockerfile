FROM golang:1.24-alpine AS builder

WORKDIR /app

# Создаем структуру директорий для правильной работы с replace директивами
COPY ./utils ./utils
COPY ./CalendarService ./CalendarService

# Переходим в директорию сервиса для сборки
WORKDIR /app/CalendarService

RUN go mod download && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o calendar_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

COPY --from=builder /app/CalendarService/calendar_service .
# Создаем папку configs и копируем туда конфигурационный файл
RUN mkdir -p configs
COPY --from=builder /app/CalendarService/configs/config.yaml ./configs/

EXPOSE 8805

CMD ["./calendar_service"] 