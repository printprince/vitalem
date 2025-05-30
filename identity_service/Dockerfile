FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o identity_service ./cmd/api

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata wget curl

WORKDIR /app

# Копируем исполняемый файл из образа сборки
COPY --from=builder /app/identity_service .
COPY --from=builder /app/config.yaml .

EXPOSE 8801

# Добавляем скрипт для проверки готовности сервисов
COPY <<EOF /app/wait-for.sh
#!/bin/sh
# wait-for.sh

set -e

host="$1"
shift
cmd="$@"

until wget --spider -q "$host"; do
  >&2 echo "Waiting for $host to be ready..."
  sleep 1
done

>&2 echo "$host is ready!"
exec $cmd
EOF

RUN chmod +x /app/wait-for.sh

# Запускаем приложение с проверкой готовности зависимостей
CMD ["/app/wait-for.sh", "http://postgres:5432", "/app/identity_service"]
