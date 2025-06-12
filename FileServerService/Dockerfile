# Этап сборки
FROM golang:1.24-alpine AS builder

# Установка необходимых зависимостей
RUN apk add --no-cache git

# Установка рабочей директории
WORKDIR /app

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование всего исходного кода
COPY . .

# Сборка бинарного файла
RUN CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/fileserver

# Финальный образ
FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование бинарного файла
COPY --from=builder /app/fileserver ./

# Копирование конфигурационного файла
COPY .env ./.env

# Открытие порта
EXPOSE 8080

# Команда запуска
CMD ["./fileserver"]
