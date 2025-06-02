FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем utils (общие компоненты)
COPY ./utils ./utils

# Копируем go.mod и go.sum
COPY ./identity_service/go.mod ./identity_service/go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY ./identity_service .

# Собираем приложение
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o identity_service ./cmd/api

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

# Копируем исполняемый файл из образа сборки
COPY --from=builder /app/identity_service .
COPY --from=builder /app/config.yaml .

EXPOSE 8801

# Запускаем приложение
CMD ["./identity_service"]
