# Report Service

Admin reporting orchestration for PayLater. **No database** — reads aggregates
from User Service and Ledger Service over REST.

## Port

Default `8085`.

## Configure

```bash
cp .env.example .env
```

| Variable | Example |
|----------|---------|
| `SERVER_PORT` | `8085` |
| `JWT_SECRET` | same as other services |
| `USER_SERVICE_URL` | `http://localhost:8082` |
| `LEDGER_SERVICE_URL` | `http://localhost:8084` |
| `INTERNAL_API_TOKEN` | must match user/ledger |

No `DB_*` variables.

## Routes

| Method | Path | Auth |
|--------|------|------|
| GET | `/health` | public |
| GET | `/admin/reports/outstanding-balance` | admin JWT |
| GET | `/admin/reports/users-due` | admin JWT |
| GET | `/admin/reports/users-at-credit-limit` | admin JWT |
| GET | `/admin/reports/merchant-commissions` | admin JWT |

## Run

```bash
go run ./cmd/server
```

Requires User (:8082) and Ledger (:8084) with the same internal token.

## Limitations / future

- No merchant-name enrichment
- No pagination
- No report DB / event-driven read model
- No Redis caching
