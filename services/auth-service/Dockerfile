# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy shared module
COPY shared ./shared

# Copy auth service
COPY services/auth-service ./services/auth-service

WORKDIR /app/services/auth-service

# Download dependencies
RUN go mod download

# Build Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/server


# Stage 2: Runtime image
FROM alpine:3.22

WORKDIR /app

# Copy only compiled binary
COPY --from=builder /auth-service .

EXPOSE 8081

CMD ["./auth-service"]
