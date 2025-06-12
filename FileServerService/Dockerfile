# Build stage
FROM golang:1.21-alpine AS builder

# Установка git
RUN apk add --no-cache git

# Установка рабочей директории
WORKDIR /app

# Копирование всего содержимого директории FileServerService
COPY . .

# Генерация go.sum и загрузка зависимостей
RUN go mod tidy && go mod download

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/fileserver

# Final stage
FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование бинарного файла из builder
COPY --from=builder /app/fileserver ./

# Открытие порта
EXPOSE 8081

# Запуск приложения
CMD ["./fileserver"]
