FROM golang:1.24-alpine AS builder

WORKDIR /app

# Создаем структуру директорий для правильной работы с replace директивами
COPY ./utils ./utils
COPY ./patient_service ./patient_service

# Переходим в директорию сервиса для сборки
WORKDIR /app/patient_service

RUN go mod download && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o patient_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/patient_service/patient_service .
COPY --from=builder /app/patient_service/config.yaml .

EXPOSE 8804

CMD ["./patient_service"]
