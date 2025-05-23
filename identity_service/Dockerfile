FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o identity_service ./cmd/api

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/identity_service .
COPY config.yaml .

EXPOSE 8801
CMD ["./identity_service"]
