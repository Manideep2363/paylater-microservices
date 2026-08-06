# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared ./shared
COPY services/report-service ./services/report-service

WORKDIR /app/services/report-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /report-service ./cmd/server


# Stage 2: Runtime
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /report-service .

EXPOSE 8085

CMD ["./report-service"]
