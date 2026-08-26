# Auth Service

Issues JWTs for users, merchants, and admin.
Phase 4: credentials are loaded/created via REST calls to user-service and merchant-service.

## Public APIs

| Method | Path |
|--------|------|
| GET | `/health` |
| POST | `/register` |
| POST | `/login` |
| POST | `/merchant/register` |
| POST | `/merchant/login` |
| POST | `/admin/login` |

## Configuration

```bash
cp .env.example .env
```

| Variable | Example |
|----------|---------|
| `SERVER_PORT` | `8081` |
| `JWT_SECRET` | same as user/merchant services |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | admin login |
| `USER_SERVICE_URL` | `http://localhost:8082` |
| `MERCHANT_SERVICE_URL` | `http://localhost:8083` |
| `INTERNAL_API_TOKEN` | must match user/merchant `INTERNAL_API_TOKEN` |

## Run with user + merchant services

Terminal 1 — user-service (`DB_NAME=paylater_users`, `INTERNAL_API_TOKEN=paylater-internal-secret`, port 8082):

```bash
cd services/user-service && go run ./cmd/server
```

Terminal 2 — merchant-service (port 8083, same internal token):

```bash
cd services/merchant-service && go run ./cmd/server
```

Terminal 3 — auth-service:

```bash
cd services/auth-service && go run ./cmd/server
```

## Demo flow

```bash
# Register user (auth -> user-service -> MySQL)
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret12"}'

# Login user (auth -> user-service lookup -> bcrypt -> JWT)
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret12"}'

# Register merchant (auth -> merchant-service -> MySQL)
curl -X POST http://localhost:8081/merchant/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Shop","email":"shop@example.com","phone":"9999999999","password":"secret12","commission_percentage":5}'

# Login merchant
curl -X POST http://localhost:8081/merchant/login \
  -H "Content-Type: application/json" \
  -d '{"email":"shop@example.com","password":"secret12"}'
```

If user/merchant services are down, auth returns **503** `{"error":"service unavailable"}`.
