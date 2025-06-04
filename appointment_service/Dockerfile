FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ./appointment_service ./appointment_service

# Build the application
RUN go mod download && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appointment_service cmd/main.go

# Переходим в директорию сервиса для сборки
WORKDIR /app/appointment_service

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary and config from builder stage
COPY --from=builder /app/appointment-service .
COPY --from=builder /app/config.yaml .

# Expose port
EXPOSE 8085

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

# Run the application
CMD ["./appointment-service"] 