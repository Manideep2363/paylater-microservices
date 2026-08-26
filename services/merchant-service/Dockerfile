# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared ./shared
COPY services/merchant-service ./services/merchant-service

WORKDIR /app/services/merchant-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /merchant-service ./cmd/server


# Stage 2: Runtime
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /merchant-service .

EXPOSE 8083

CMD ["./merchant-service"]
