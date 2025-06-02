FROM golang:1.24-alpine AS builder

WORKDIR /app

# Создаем структуру директорий для правильной работы с replace директивами
COPY ./utils ./utils
COPY ./identity_service ./identity_service

# Переходим в директорию сервиса для сборки
WORKDIR /app/identity_service

# Загружаем зависимости и собираем приложение
RUN go mod download && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o identity_service ./cmd/api

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

# Копируем исполняемый файл из образа сборки
COPY --from=builder /app/identity_service/identity_service .
COPY --from=builder /app/identity_service/config.yaml .

EXPOSE 8801

# Запускаем приложение
CMD ["./identity_service"]
