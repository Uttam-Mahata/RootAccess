# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**RootAccess** is a full-stack CTF (Capture The Flag) platform. Backend is Go/Gin, frontend is Angular 21. Data is stored in MongoDB with Redis for caching.

## Development Commands

### Infrastructure (required first)
```bash
docker compose up -d   # Starts MongoDB on :27017
```

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
ng test         # Runs Angular tests
```

### Swagger / API Docs
```bash
# Install swag CLI once:
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate docs (run from repo root or backend/):
cd backend && swag init -g cmd/api/main.go
```
Swagger UI is available at `http://localhost:8080/swagger/index.html` in development. It is **disabled in production** via the `-tags production` build flag.

### Client SDK Generation
```bash
./scripts/generate-clients.sh  # Generates TypeScript + Python clients into /clients/
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

```
backend/
  cmd/api/main.go          # Entry point: loads config, connects DB/Redis, starts Gin router
  cmd/admin/main.go        # Interactive CLI for admin user management
  internal/
    config/config.go       # Env-based config (godotenv); all secrets here
    database/              # MongoDB (global DB var) + Redis (global RDB var) connections
    models/                # Data structs + business logic (e.g., dynamic scoring in challenge.go)
    repositories/          # MongoDB CRUD — one file per collection
    services/              # Business logic, orchestrates repositories
    handlers/              # Gin HTTP controllers, one per domain
    middleware/            # Auth (JWT cookie → header fallback), AdminOnly, RateLimit, Audit
    routes/routes.go       # Single file that wires all repositories → services → handlers → routes
    websocket/hub.go       # WebSocket hub for real-time scoreboard/flag solve events
```

**Dependency flow:** `routes.go` instantiates all repos, services, and handlers, then registers routes. There is no DI framework — constructor injection is used manually.

**Authentication:** JWT stored in HTTP-only cookie (`auth_token`). Middleware also accepts `Authorization: Bearer <token>` header as fallback. Admin access is a separate `AdminMiddleware` that checks the `role` claim.

**Scoring:** Dynamic scoring (CTFd formula) is implemented in `internal/models/challenge.go`. Points decay as solve count increases, clamped between `MinPoints` and `MaxPoints`.

**Caching:** Scoreboard data is cached in Redis. Cache keys include a freeze-time suffix when the contest scoreboard is frozen.

**WebSocket:** `internal/websocket/hub.go` broadcasts events (e.g., first blood, flag solves) to connected clients in real time.

### Frontend — Angular 21

```
frontend/src/app/
  app.routes.ts            # Route definitions with authGuard, adminGuard, guestGuard
  app.config.ts            # App-level providers
  components/              # One directory per page/feature
  services/                # API client services (one per backend domain)
  interceptors/            # HTTP interceptors (e.g., credentials: 'include' for cookies)
```

Route guards (`authGuard`, `adminGuard`, `guestGuard`) wait for `AuthService.authCheckComplete$` before navigating, preventing race conditions on page load.

### Environment Variables (Backend)

Key variables (see `backend/.env.example` for full list):

| Variable | Purpose |
|---|---|
| `MONGO_URI` | MongoDB connection string |
| `DB_NAME` | Database name (default: `go_ctf`) |
| `JWT_SECRET` | Must be 32+ random characters in production |
| `REDIS_ADDR` | Redis address (default: `localhost:6379`) |
| `APP_ENV` | Set to `production` to disable Swagger |
| `TRUSTED_PROXIES` | Comma-separated proxy IPs for correct client IP |
| `SMTP_*` | Email config for verification and password reset |
| `GOOGLE/GITHUB/DISCORD_*` | OAuth provider credentials |

## Key Conventions

- **Role promotion:** Default registration role is hardcoded to `"user"`. Promote to admin only via the CLI tool or direct MongoDB update — never via API.
- **Swagger annotations:** Swagger metadata lives in `cmd/api/main.go` (title, version, host). Handler-level annotations are in each handler file.
- **Build tags:** The `registerSwagger(r)` call in `routes.go` is conditionally compiled — there is a `swagger_dev.go` (development) and a no-op `swagger_prod.go` (built with `-tags production`).
- **Rate limiting:** Flag submission is limited to 5 attempts/minute per challenge. Auth endpoints are limited per IP.
- **Audit logging:** All admin route mutations go through `AuditMiddleware`, which records actions to the `audit_logs` collection.

## CI/CD & Releases

Pushing a tag matching `v*` triggers `.github/workflows/release.yml`, which:
1. Generates Swagger docs
2. Runs `scripts/generate-clients.sh` to build TS and Python clients
3. Publishes to NPM (`@rootaccessd/...`) and PyPI in parallel
4. Creates a GitHub Release with `swagger.json` and client zips attached
