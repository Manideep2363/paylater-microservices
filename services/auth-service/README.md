# Auth Service

Issues JWTs for users, merchants, and admin. Phase 1 uses in-memory credential stores until user-service and merchant-service exist.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness |
| POST | `/register` | User register |
| POST | `/login` | User login → `{ "token": "..." }` |
| POST | `/merchant/register` | Merchant register |
| POST | `/merchant/login` | Merchant login → `{ "token": "..." }` |
| POST | `/admin/login` | Admin login (env credentials) → `{ "token": "..." }` |

## Run

```bash
cd services/auth-service
# optional .env: SERVER_PORT, JWT_SECRET, ADMIN_EMAIL, ADMIN_PASSWORD
go run ./cmd/server
```

Default port: `8080` (override with `SERVER_PORT`).
