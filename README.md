# PayLater Microservices

Buy-now-pay-later platform migrated from a modular monolith (`../paylater`) into independently runnable Go microservices (Go 1.25+, Gin) in a monorepo.

Clients should talk to the **API Gateway** (`:8080`). Downstream services own domains and databases. Auth is JWT (validated by each service). Service-to-service calls use `X-Internal-Token`.

## Architecture

```text
Client (Postman / curl)
        │
        ▼
  api-gateway :8080          ← public allowlist; strips X-Internal-Token; blocks /internal/*
        │
        ├─► auth-service :8081          (no DB)     → user + merchant (S2S)
        ├─► user-service :8082          paylater_users
        ├─► merchant-service :8083      paylater_merchants
        ├─► ledger-service :8084        paylater_ledger  → user + merchant (S2S)
        └─► report-service :8085        (no DB)          → user + ledger (S2S)
```

| Service | Port | Database | Role |
|---------|------|----------|------|
| **api-gateway** | 8080 | — | Public reverse-proxy allowlist |
| **auth-service** | 8081 | — | Register/login; issues JWTs |
| **user-service** | 8082 | `paylater_users` | Profiles, credit limit, sole owner of `current_due` |
| **merchant-service** | 8083 | `paylater_merchants` | Profiles, commission % (3–20) |
| **ledger-service** | 8084 | `paylater_ledger` | Purchases & repayments; due compensation |
| **report-service** | 8085 | — | Admin report BFF |

Roles: `user` | `merchant` | `admin`. Admin credentials come from env (`ADMIN_EMAIL` / `ADMIN_PASSWORD`); JWT `user_id=0`.

## Folder structure

```text
paylater-microservices/
├── services/
│   ├── api-gateway/
│   ├── auth-service/
│   ├── user-service/
│   ├── merchant-service/
│   ├── ledger-service/
│   └── report-service/
├── shared/                 # Go replace → ../../shared
│   ├── auth/               # JWT HS256 + bcrypt
│   ├── middleware/         # AuthMiddleware, RequireRole
│   ├── config/             # Env loading
│   ├── response/           # JSON helpers
│   └── httpclient/         # S2S REST client
├── docker/
│   └── mysql/init.sql      # Creates DBs + tables on first MySQL boot
├── docs/                   # Architecture + ontology
├── docker-compose.yml
├── Micro_services.postman_collection.json
├── .env.example
└── README.md
```

Each service has its own `go.mod`, `cmd/server`, and (where applicable) sqlc + MySQL schema.

## Quick start (Docker Compose)

Requirements: Docker Desktop / Docker Compose.

```bash
cp .env.example .env
# Edit .env: MYSQL_ROOT_PASSWORD, JWT_SECRET, INTERNAL_API_TOKEN, ADMIN_*

docker compose up -d
```

| Endpoint | URL |
|----------|-----|
| API Gateway | http://localhost:8080 |
| MySQL (host) | localhost:**3307** → container 3306 |
| Auth / User / Merchant / Ledger / Report | 8081–8085 (also exposed) |

Health check:

```bash
curl http://localhost:8080/health
```

Compose pulls prebuilt images (`manideep2363/paylater-*-service:v1`) and starts MySQL with `docker/mysql/init.sql`.

Import `Micro_services.postman_collection.json` and target **http://localhost:8080**.

Stop:

```bash
docker compose down
```

## Run locally (Go)

Requirements: Go 1.25+, MySQL with the three databases created (see `docker/mysql/init.sql` or per-service `sql/schema.sql`).

1. Start MySQL (or `docker compose up -d mysql`).
2. Copy each service’s `.env.example` → `.env` and align secrets:
   - Same `JWT_SECRET` everywhere JWT is validated
   - Same `INTERNAL_API_TOKEN` on auth, user, merchant, ledger, report
3. Start services (separate terminals), then the gateway:

```bash
cd services/user-service && go run ./cmd/server      # :8082
cd services/merchant-service && go run ./cmd/server  # :8083
cd services/auth-service && go run ./cmd/server      # :8081
cd services/ledger-service && go run ./cmd/server    # :8084
cd services/report-service && go run ./cmd/server    # :8085
cd services/api-gateway && go run ./cmd/server       # :8080
```

Use **http://localhost:8080** from clients. Do not call `/internal/*` through the gateway (404).

## Core flows

1. **Register / login** — auth → user or merchant internal APIs → JWT  
2. **Purchase** — ledger → merchant (commission) → user IncreaseDue → INSERT transaction; on insert failure Compensate DecreaseDue  
3. **Repayment** — ledger → user DecreaseDue → INSERT payment; on insert failure Compensate IncreaseDue  
4. **Admin reports** — report → user / ledger internal report APIs  

## Shared library

Imported via Go `replace` to `../../shared`:

- **auth** — JWT generate/validate (24h), bcrypt  
- **middleware** — Bearer JWT + role checks  
- **config** — env / `.env` loading  
- **response** — `{"error"}` / `{"message"}` helpers  
- **httpclient** — JSON S2S client with default headers (e.g. internal token)

## Documentation

| Doc | Description |
|-----|-------------|
| [docs/PROJECT_ANALYSIS.md](docs/PROJECT_ANALYSIS.md) | Full system analysis, APIs, sequences, gaps |
| [docs/ONTOLOGY.md](docs/ONTOLOGY.md) | Domain ontology, glossary, cheat sheet |
| [services/*/README.md](services/) | Per-service setup and endpoints |

## Known limitations

- REST compensation is not fully idempotent; no reconciliation worker  
- API amounts use `float64`; MySQL uses `DECIMAL` (ledger normalizes to 2dp)  
- Report service: no merchant-name enrichment, pagination, or caching  
- Gateway CORS not configured (Postman/curl oriented)  

## License / monolith

The original monolith remains at `../paylater` and is unchanged by this repo.
