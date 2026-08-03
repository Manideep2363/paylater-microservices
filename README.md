# PayLater Microservices

Monorepo scaffold for migrating the PayLater modular monolith into independently runnable Go microservices.

**Phase 0 only:** project structure, shared libraries, and health-check entrypoints. No business logic, SQLC, or databases yet.

The original monolith remains at `../paylater` and is unchanged.

## Folder structure

```text
paylater-microservices/
├── services/           # Independently runnable Go services (each has its own go.mod)
│   ├── auth-service/
│   ├── user-service/
│   ├── merchant-service/
│   ├── ledger-service/
│   └── report-service/
├── shared/             # Reusable libraries imported by services via replace
│   ├── auth/           # JWT + bcrypt
│   ├── middleware/     # AuthMiddleware, RequireRole
│   ├── config/         # Environment loading
│   ├── response/       # Common JSON helpers
│   └── httpclient/     # Placeholder for future REST S2S calls
├── docs/               # Architecture and migration notes
├── scripts/            # Helper scripts (later)
├── docker/             # Docker / Compose assets (later)
├── README.md
└── .gitignore
```

## Purpose of each service

| Service | Responsibility (planned) |
|---------|--------------------------|
| **auth-service** | Register/login for users, merchants, and admin; issue JWTs |
| **user-service** | User profiles, credit limit, sole owner of `current_due` |
| **merchant-service** | Merchant profiles and commission percentage |
| **ledger-service** | Purchases and repayments (transactions + payments) |
| **report-service** | Admin reports (outstanding dues, credit-limit users, commissions) |

Each service currently only exposes `GET /health`.

## Purpose of `shared/`

Code that every service needs without owning a domain:

- **auth** — JWT generate/validate and bcrypt helpers (moved from the monolith)
- **middleware** — Bearer JWT validation and role checks
- **config** — load env vars (optional `.env`)
- **response** — consistent JSON error/message helpers
- **httpclient** — thin HTTP client stub for later service-to-service REST

Services depend on shared with a Go `replace` directive pointing at `../../shared`.

## How to run one service

Requirements: Go 1.25+

```bash
cd services/user-service
go mod tidy
go run ./cmd/server
```

Then:

```bash
curl http://localhost:8080/health
```

Expected:

```json
{"service":"user-service","status":"ok"}
```

Override the port:

```bash
# Windows PowerShell
$env:SERVER_PORT="8082"; go run ./cmd/server

# bash
SERVER_PORT=8082 go run ./cmd/server
```

Suggested local ports when running several services later:

| Service | Port |
|---------|------|
| auth-service | 8081 |
| user-service | 8082 |
| merchant-service | 8083 |
| ledger-service | 8084 |
| report-service | 8085 |

## What is intentionally not done yet

- No Auth / User / Merchant / Ledger / Report business APIs
- No MySQL connections or SQL schemas
- No SQLC
- No service-to-service REST
- No Docker Compose

Wait for approval before Phase 1.
