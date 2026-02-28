# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**RootAccess** is a full-stack CTF (Capture The Flag) platform. Backend is Go/Gin, frontend is Angular 21. Data is stored in MongoDB with Redis for caching.

## Development Commands

### Backend
```bash
cd backend
cp .env.example .env   # First time only — fill in secrets
go mod download
go run cmd/api/main.go  # Runs on :8080
```

Run all backend tests:
```bash
cd backend && go test ./...
```

Run a single test file or package:
```bash
cd backend && go test ./internal/models/... -run TestCurrentPoints_DynamicScoring -v
```

### Frontend
```bash
cd frontend
npm install
npm start       # Dev server on :4200
npm run build   # Production build
ng test         # Runs Angular tests (Vitest + Playwright also available)
```

### Swagger / API Docs
```bash
# Install swag CLI once:
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate docs (run from repo root or backend/):
cd backend && swag init -g cmd/api/main.go
```

### Admin CLI Tool
```bash
# In Docker:
docker exec -it go_ctf_backend ./admin-tool
# Or build locally:
cd backend && go run cmd/admin/main.go
```

## Architecture

### Backend — Clean Architecture layers

Go module path: `github.com/Uttam-Mahata/RootAccess/backend`

```
backend/
  cmd/api/main.go          # Entry point: loads config, connects DB/Redis, starts Gin router
  cmd/admin/main.go        # Interactive CLI for admin user management
  internal/
    config/config.go       # Env-based config (godotenv + optional AWS Secrets Manager)
    database/              # MongoDB (global DB var) + Redis (global RDB var) connections
    models/                # Data structs + business logic (e.g., dynamic scoring in challenge.go)
    repositories/          # MongoDB CRUD — one file per collection
    services/              # Business logic, orchestrates repositories
    handlers/              # Gin HTTP controllers, one per domain
    middleware/            # Auth, AdminOnly, RateLimit, IPRateLimit, Audit, Contest
    routes/routes.go       # Single file that wires all repos → services → handlers → routes
    utils/                 # crypto.go, errors.go — shared utilities
    websocket/hub.go       # Hub interface + 3 implementations (see below)
```

**Dependency flow:** `routes.go` instantiates all repos, services, and handlers, then registers routes. No DI framework — constructor injection only.

**Authentication:** JWT stored in HTTP-only cookie (`auth_token`). Middleware also accepts `Authorization: Bearer <token>` header as fallback. Admin access uses a separate `AdminMiddleware` checking the `role` claim.

**Scoring:** Dynamic scoring (CTFd formula) in `internal/models/challenge.go`. Points decay as solve count increases, clamped between `MinPoints` and `MaxPoints`.

**Caching:** Scoreboard data is cached in Redis. Cache keys include a freeze-time suffix when the contest scoreboard is frozen.

**WebSocket — 3 Hub implementations** (auto-selected in `routes.go`):
- `MemoryHub` — in-process; used for local dev and single-node deployments
- `RedisHub` — pub/sub via Redis; used for multi-instance deployments (Redis present, no Lambda)
- `AwsLambdaHub` — uses API Gateway Management API; used when `AWS_LAMBDA_FUNCTION_NAME` env var is set

**Contest system:** Beyond the basic global scoreboard there is a multi-entity contest system with `ContestEntity → ContestRound → RoundChallenge` and `TeamContestRegistration` for per-contest team sign-ups. Managed via `/admin/contest-entities/*` routes.

### Frontend — Angular 21

```
frontend/src/app/
  app.routes.ts            # Route definitions — all components are lazy-loaded standalone
  app.config.ts            # App-level providers
  components/              # One directory per page/feature
  services/                # API client services (one per backend domain)
  interceptors/            # credentials.interceptor.ts — adds withCredentials: true to all requests
```

Route guards (`authGuard`, `adminGuard`, `guestGuard`, `landingGuard`) wait for `AuthService.authCheckComplete$` before navigating, preventing race conditions on page load.

### Environment Variables (Backend)

Key variables (see `backend/.env.example` for full list):

| Variable | Purpose |
|---|---|
| `MONGO_URI` | MongoDB connection string |
| `DB_NAME` | Database name (default: `go_ctf`) |
| `JWT_SECRET` | Must be 32+ random characters in production |
| `REDIS_ADDR` | Redis address (default: `localhost:6379`) |
| `APP_ENV` | Set to `production` to disable Swagger |
| `FRONTEND_URL` | Used for CORS and OAuth redirect URLs |
| `TRUSTED_PROXIES` | Comma-separated proxy IPs for correct client IP |
| `CORS_ALLOWED_ORIGINS` | Additional allowed CORS origins (comma-separated) |
| `SMTP_*` | Email config for verification and password reset |
| `GOOGLE/GITHUB/DISCORD_*` | OAuth provider credentials |
| `WS_CALLBACK_URL` | AWS API Gateway WebSocket callback URL (Lambda mode) |
| `AWS_SECRET_NAME` | If set, loads all secrets from AWS Secrets Manager instead of `.env` |
| `REGISTRATION_MODE` | `open` (default), `domain` (email domain allowlist), or `disabled` |
| `REGISTRATION_ALLOWED_DOMAINS` | Comma-separated domains for `domain` registration mode |

## Key Conventions

- **Role promotion:** Default registration role is hardcoded to `"user"`. Promote to admin only via the CLI tool or direct MongoDB update — never via API.
- **Swagger annotations:** Swagger metadata lives in `cmd/api/main.go` (title, version, host). Handler-level annotations are in each handler file.
- **Build tags:** `registerSwagger(r)` in `routes.go` is conditionally compiled — `swagger_dev.go` (development) and a no-op `swagger_prod.go` (built with `-tags production`).
- **Rate limiting:** Flag submission is limited to 5 attempts/minute per challenge. Auth endpoints are limited 10/minute per IP. Use `middleware.RateLimitMiddleware` (per-user) vs `middleware.IPRateLimitMiddleware` (per-IP).
- **Audit logging:** All admin route mutations go through `AuditMiddleware`, which records actions to the `audit_logs` collection.
- **CORS:** Allowed origins are hardcoded in `routes.go` plus `CORS_ALLOWED_ORIGINS` env var. In `development` environment, all origins are allowed.

## CI/CD & Releases

- `.github/workflows/backend-deploy.yml` — deploys backend on push
- `.github/workflows/frontend-deploy.yml` — deploys frontend on push
