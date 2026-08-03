# User Service

Owns the `users` table in the `paylater_users` MySQL database.
Phase 2: independently runnable; no service-to-service REST yet.

## Responsibilities

- Create / get / list users
- Atomic `IncreaseDue` / `DecreaseDue` (credit rules from the monolith)
- Internal APIs for future auth / ledger callers

## Setup database

```sql
CREATE DATABASE paylater_users;
USE paylater_users;
SOURCE sql/schema.sql;
```

Or from this directory:

```bash
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS paylater_users;"
mysql -u root -p paylater_users < sql/schema.sql
```

## Configure

Copy `.env.example` to `.env` and set values:

```bash
cp .env.example .env
```

Required: `DB_*`, `JWT_SECRET`, `SERVER_PORT` (default 8082 if unset).

Required for `/internal/*`: `INTERNAL_API_TOKEN` — callers must send header `X-Internal-Token`.
If unset/empty, all `/internal/*` requests are rejected (fail closed).

## Generate SQLC / run

```bash
sqlc generate
go mod tidy
go run ./cmd/server
```

## Endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/health` | none |
| GET | `/users/:id` | JWT (self or admin) |
| GET | `/admin/users` | JWT admin |
| POST | `/admin/users` | JWT admin |
| POST | `/internal/users` | `X-Internal-Token` required |
| GET | `/internal/users/:id` | `X-Internal-Token` required |
| GET | `/internal/users/by-email?email=` | `X-Internal-Token` required |
| POST | `/internal/users/:id/due/increase` | `X-Internal-Token` required |
| POST | `/internal/users/:id/due/decrease` | `X-Internal-Token` required |

Public/admin responses never include password hashes. Internal get/create responses include the password hash for future auth-service use.

## Smoke test examples

See README section in the Phase 2 summary, or:

```bash
# health
curl http://localhost:8082/health

# admin token from auth-service (or any JWT with role=admin + same JWT_SECRET)
curl -X POST http://localhost:8082/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret12"}'
```
