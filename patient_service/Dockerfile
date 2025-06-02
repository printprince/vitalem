FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем utils (общие компоненты)
COPY ./utils ./utils

COPY ./patient_service/go.mod ./patient_service/go.sum ./

RUN go mod download

COPY ./patient_service .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o patient_service ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/patient_service .
COPY --from=builder /app/config.yaml .

EXPOSE 8804

CMD ["./patient_service"]
