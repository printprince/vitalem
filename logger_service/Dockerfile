# Logger Service Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o logger_service ./cmd/api

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/logger_service .
COPY config.yaml .

EXPOSE 8802
CMD ["./logger_service"]