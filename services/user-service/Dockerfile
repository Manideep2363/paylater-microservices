# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Shared Go module
COPY shared ./shared

# User service
COPY services/user-service ./services/user-service

WORKDIR /app/services/user-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/server


# Stage 2: Runtime
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /user-service .

EXPOSE 8082

CMD ["./user-service"]
