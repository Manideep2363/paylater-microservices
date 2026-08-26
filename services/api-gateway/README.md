# API Gateway

Public entry point for PayLater microservices. Lightweight Gin + `httputil.ReverseProxy`.

**No** JWT validation, roles, databases, or `INTERNAL_API_TOKEN`.

## Port

Default `8080`.

## Configure

```bash
cp .env.example .env
```

| Variable | Default |
|----------|---------|
| `SERVER_PORT` | `8080` |
| `AUTH_SERVICE_URL` | `http://localhost:8081` |
| `USER_SERVICE_URL` | `http://localhost:8082` |
| `MERCHANT_SERVICE_URL` | `http://localhost:8083` |
| `LEDGER_SERVICE_URL` | `http://localhost:8084` |
| `REPORT_SERVICE_URL` | `http://localhost:8085` |

Not used: `JWT_SECRET`, `INTERNAL_API_TOKEN`, `DB_*`.

## Security

- Explicit public route allowlist only
- `/internal/*` blocked (404)
- Client `X-Internal-Token` stripped before proxy
- Downstream services still validate JWT/roles

## CORS

Gin middleware allows the Vite React app in development:

| Variable | Default |
|----------|---------|
| `CORS_ALLOWED_ORIGIN` | `http://localhost:5173` |

Allowed methods: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`.  
Allowed headers: `Content-Type`, `Authorization`, `X-Request-ID`.  
OPTIONS preflight is answered by the gateway (204) and is not proxied.

**Production:** set `CORS_ALLOWED_ORIGIN` to your real frontend origin (e.g. `https://app.example.com`). Never use `*`.

## Run with all services

Start Auth, User, Merchant, Ledger, Report, then:

```bash
cd services/api-gateway
go run ./cmd/server
```

Use **only** `http://localhost:8080` from clients.
