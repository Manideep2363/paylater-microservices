# Merchant Service

Owns the `merchants` table in the `paylater_merchants` MySQL database.
Phase 3: independently runnable; not wired to auth/ledger yet.

## Responsibilities

- Create / get / list merchants
- Merchant self-profile (`GET /merchant/profile`)
- Update commission (3%–20%)
- Internal APIs for future auth-service and ledger-service

## Setup database

```sql
CREATE DATABASE paylater_merchants;
USE paylater_merchants;
SOURCE sql/schema.sql;
```

Or:

```bash
mysql -u root -p < scripts/init_db.sql
```

## Configure

```bash
cp .env.example .env
```

| Variable | Purpose |
|----------|---------|
| `SERVER_PORT` | Default `8083` if unset |
| `DB_*` | MySQL connection; default DB name `paylater_merchants` |
| `JWT_SECRET` | Must match token issuer (auth-service) |
| `INTERNAL_API_TOKEN` | **Required** for `/internal/*` (fail closed) |

### Internal authentication

Every `/internal/*` route requires:

1. Non-empty `INTERNAL_API_TOKEN` in the environment (otherwise all internal calls get 401)
2. Request header `X-Internal-Token: <same value>`

Missing or wrong tokens are rejected with 401.

## Run

```bash
sqlc generate
go mod tidy
go run ./cmd/server
```

## Endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/health` | none |
| GET | `/merchant/profile` | JWT role `merchant` (ID from token) |
| POST | `/admin/merchants` | JWT admin |
| GET | `/admin/merchants` | JWT admin |
| GET | `/admin/merchants/:id` | JWT admin |
| PUT | `/admin/merchants/:id/commission` | JWT admin |
| POST | `/internal/merchants` | `X-Internal-Token` |
| GET | `/internal/merchants/:id` | `X-Internal-Token` |
| GET | `/internal/merchants/by-email?email=` | `X-Internal-Token` |

Public/admin responses never include `password_hash`. Internal get/create responses include `password_hash` for future auth use.

Commission must be between **3 and 20** inclusive (same rule as the monolith).
