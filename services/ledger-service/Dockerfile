# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared ./shared
COPY services/ledger-service ./services/ledger-service

WORKDIR /app/services/ledger-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /ledger-service ./cmd/server


# Stage 2: Runtime
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /ledger-service .

EXPOSE 8084

CMD ["./ledger-service"]
