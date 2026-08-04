# Ledger Service

Owns purchases (`transactions`) and repayments (`payments`) in `paylater_ledger`.
Coordinates credit via User Service and commission via Merchant Service over REST.

**Does not** access `paylater_users` or `paylater_merchants` databases.

## Money note (technical debt)

Amounts use MySQL `DECIMAL` and are normalized to **two decimal places** in Go.
API layers still use `float64` for compatibility with user/merchant services.
A consistent decimal / minor-unit representation should be introduced across
services in a later hardening phase.

## Future hardening (not in Phase 5)

- End-to-end operation idempotency (Ledger-only `Idempotency-Key` is insufficient
  when User Service commits a due update but Ledger times out)
- Reconciliation workers for failed compensation
- Stronger money types

## Setup database

```bash
mysql -u root -p < scripts/init_db.sql
```

## Configure

```bash
cp .env.example .env
```

| Variable | Default / example |
|----------|-------------------|
| `SERVER_PORT` | `8084` |
| `DB_NAME` | `paylater_ledger` |
| `JWT_SECRET` | same as auth/user/merchant |
| `USER_SERVICE_URL` | `http://localhost:8082` |
| `MERCHANT_SERVICE_URL` | `http://localhost:8083` |
| `INTERNAL_API_TOKEN` | must match user/merchant |

## Run with Auth + User + Merchant

```bash
# user :8082, merchant :8083, auth :8081, then:
cd services/ledger-service
go run ./cmd/server
```

## Endpoints

| Method | Path | Role |
|--------|------|------|
| GET | `/health` | public |
| POST | `/purchases` | user (JWT id) |
| POST | `/payments` | user (JWT id) |
| GET | `/payments` | user (own) |
| GET | `/admin/purchases` | admin |
| GET | `/admin/purchases/:id` | admin |
| GET | `/admin/users/:id/purchases` | admin |
| GET | `/admin/payments/:id` | admin |
| GET | `/admin/users/:id/payments` | admin |
| GET | `/merchant/transactions` | merchant (JWT id) |

## Compensation

Purchase: IncreaseDue → INSERT tx; on INSERT failure → DecreaseDue(same amount).  
Repayment: DecreaseDue → INSERT payment; on INSERT failure → IncreaseDue(same amount).  
Compensation failure → HTTP 500 + structured critical log (no internal details to client).
