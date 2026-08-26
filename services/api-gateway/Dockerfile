
# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy shared module
COPY shared ./shared

# Copy API Gateway
COPY services/api-gateway ./services/api-gateway

WORKDIR /app/services/api-gateway

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/server


# Stage 2: Runtime
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
