# Go CTF Platform - Developer Context

## 1. Project Overview
**Name:** RootAccess (Go CTF Platform)
**Type:** Full-stack Web Application (Capture The Flag)
**Repository:** `github.com/Uttam-Mahata/RootAccess`
**Goal:** Provide a platform for hosting and solving security challenges, with user management, scoring, and admin capabilities.

## 2. Technology Stack

### Backend (`/backend`)
- **Language:** Go 1.24
- **Module Path:** `github.com/Uttam-Mahata/RootAccess/backend`
- **Framework:** Gin (HTTP)
- **Documentation:** Swagger/OpenAPI (via `swaggo`)
- **Database Driver:** `go.mongodb.org/mongo-driver`
- **Authentication:** `golang-jwt/jwt` (JWT)
- **Architecture:** Clean Architecture
  - `cmd/api`: Entry point
  - `internal/handlers`: HTTP Controllers
  - `internal/services`: Business Logic
  - `internal/repositories`: Database Access
  - `internal/models`: Data Structures

### Frontend (`/frontend`)
- **Framework:** Angular 21
- **UI Library:** Angular Material
- **Styling:** SCSS & TailwindCSS (`@tailwindcss/postcss` detected)
- **Build Tool:** Angular CLI
- **Testing:** Vitest & Playwright

### Infrastructure
- **Database:** MongoDB 7.0 (via Docker)
- **Caching:** Redis (for state and scoreboard caching)
- **Orchestration:** Docker Compose

## 3. Development Setup & Commands

### Prerequisites
- Go 1.24+
- Node.js 18+ & npm 11+
- Docker & Docker Compose
- `swag` CLI (for API docs): `go install github.com/swaggo/swag/cmd/swag@latest`

### Quick Start
1. **Start Services:** `docker compose up -d`
2. **Backend:** `cd backend && go run cmd/api/main.go`
3. **Frontend:** `cd frontend && npm install && npm start`

### API Documentation (Swagger)
- **Development:** Accessible at `http://localhost:8080/swagger/index.html`.
- **Update Docs:** `cd backend && swag init -g cmd/api/main.go`.
- **Production:** Swagger is **disabled** in production builds (via `-tags production`).

### Client SDK Generation
Automated clients can be generated from the Swagger spec:
- **Script:** `./scripts/generate-clients.sh`
- **Output:** `/clients/typescript` and `/clients/python`.

## 4. Key Conventions & Architecture

- **Environment Management:**
  - `APP_ENV`: Set to `production` to disable Swagger and enable production optimizations.
  - Backend uses `godotenv` to load `.env`.

- **CI/CD & Releases:**
  - Pushing a version tag (e.g., `v1.0.0`) triggers a GitHub Action (`.github/workflows/release.yml`).
  - Automatically generates Swagger docs, packages TS/Python clients, and creates a GitHub Release.

- **Role Management:**
  - Default registration role is **hardcoded** to "user".
  - Admin promotion must be done via the CLI tool (`cmd/admin/main.go`).

- **Agent Skills:**
  - `api-manager`: Specialized skill for syncing Swagger docs and refreshing client SDKs. Activate by asking to "update docs" or "sync API".

## 5. Important Files
- `backend/cmd/api/main.go`: Backend entry point & Swagger metadata.
- `backend/internal/routes/routes.go`: Router setup (with build-tag based Swagger registration).
- `scripts/generate-clients.sh`: Multi-language SDK generation script.
- `.github/workflows/release.yml`: Automated release pipeline.
- `docker-compose.yml`: MongoDB and local service configuration.
