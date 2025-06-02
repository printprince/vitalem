FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем utils (общие компоненты)
COPY ./utils ./utils

COPY ./specialist_service/go.mod ./specialist_service/go.sum ./

RUN go mod download

COPY ./specialist_service .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o specialist_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget curl postgresql-client

WORKDIR /app

COPY --from=builder /app/specialist_service .
COPY --from=builder /app/config.yaml .

EXPOSE 8803

CMD ["./specialist_service"]
