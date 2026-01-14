# Go CTF Platform - Developer Context

## 1. Project Overview
**Name:** Go CTF Platform
**Type:** Full-stack Web Application (Capture The Flag)
**Goal:** Provide a platform for hosting and solving security challenges, with user management, scoring, and admin capabilities.

## 2. Technology Stack

### Backend (`/backend`)
- **Language:** Go 1.24
- **Framework:** Gin (HTTP)
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
- **Orchestration:** Docker Compose


## 3. Development Setup & Commands

### Prerequisites
- Go 1.24+
- Node.js 18+ & npm 11+
- Docker & Docker Compose
- MongoDB (if not using Docker)

### Quick Start
1. **Start Database:**
   ```bash
   docker compose up -d
   ```

2. **Backend Setup:**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env if necessary (default: PORT=8080, MONGO_URI=mongodb://localhost:27017)
   go mod download
   go run cmd/api/main.go
   ```

3. **Frontend Setup:**
   ```bash
   cd frontend
   npm install
   npm start
   # Access at http://localhost:4200
   ```

### Administrative Actions
**Create Admin User:**
The platform protects admin creation. Use the CLI tool:
```bash
cd backend
go run cmd/admin/main.go
# Follow interactive prompts to create or promote a user
```

### Testing
- **Backend:** `cd backend && go test ./...`
- **Frontend:** `cd frontend && npm test`

## 4. Key Conventions & Architecture

- **Role Management:**
  - Default registration role is **hardcoded** to "user".
  - Admin promotion must be done via the CLI tool (`cmd/admin/main.go`) or direct DB access.
  - Role checks are implemented in `internal/middleware/auth_middleware.go`.

- **Configuration:**
  - Backend uses `godotenv` to load `.env`.
  - Config struct defined in `internal/config/config.go`.

- **Frontend Structure:**
  - `src/app/components`: Feature-based components (login, scoreboard, etc.)
  - `src/app/services`: API communication (auth, challenge, etc.)
  - `src/app/interceptors`: HTTP interceptors for JWT injection (`credentials.interceptor.ts`).

## 5. Important Files
- `backend/cmd/api/main.go`: Backend entry point.
- `backend/cmd/admin/main.go`: Admin management CLI.
- `backend/internal/routes/routes.go`: API Route definitions.
- `frontend/src/app/app.routes.ts`: Frontend routing.
- `docker-compose.yml`: MongoDB service config.
