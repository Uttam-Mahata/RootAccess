# RootAccess CTF Platform

A high-performance, full-stack Capture The Flag (CTF) platform built with Go (Gin) for the backend and Angular (v21) for the frontend. Designed for scalability with Redis caching and Turso DB (LibSQL) for high-performance data storage.

## 🚀 Features

- **Dynamic Scoring**: Points for challenges decrease as more teams solve them (CTFd formula).
- **Team-Based Competition**: Create or join teams to solve challenges and climb the leaderboard together.
- **Real-time Scoreboard**: Cached global and team rankings.
- **Admin Management**: Dedicated dashboard for challenge creation, notification broadcasts, and user moderation.
- **Robust Security**: 
  - JWT authentication with HTTP-only cookies.
  - Rate limiting on flag submissions.
  - Email verification and secure password reset.
  - Role-based access control (RBAC).
- **Performance Optimized**: 
  - **Redis Caching**: Frequently accessed data like the scoreboard is cached in-memory.
  - **Turso DB (LibSQL)**: Edge database integration ensuring fast global reads and high concurrency.

## 🏗️ Architecture Diagram
 [![Architecture Diagram](https://raw.githubusercontent.com/Uttam-Mahata/RootAccess/main/rootaccess-aws.png)](https://raw.githubusercontent.com/Uttam-Mahata/RootAccess/main/rootaccess-aws.png)


### Backend
- **Language**: Go 1.24
- **Framework**: Gin (HTTP web framework)
- **Primary Database**: Turso DB (LibSQL)
- **Cache**: Redis 7.x
- **Email**: SMTP integration for verification and resets.

### Frontend
- **Framework**: Angular 21
- **Styling**: Tailwind CSS v4 & SCSS
- **UX/UI**: Material Design principles with custom dark/light theme support.

## 📋 Prerequisites

- **Docker & Docker Compose** (Recommended for production)
- **Go**: Version 1.24+ (For local development)
- **Node.js**: Version 22+ (For local development)
- **Turso CLI**: Optional (For DB management)
- **Redis**: Version 6.0+

## 🛠️ Setup Instructions

### Production Deployment (Docker)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Uttam-Mahata/RootAccess.git
   cd RootAccess
   ```

2. **Configure Environment:**
   Copy the example compose file and update your credentials:
   ```bash
   cp docker-compose.prod.example.yml docker-compose.prod.yml
   # Edit docker-compose.prod.yml with your SMTP, DB, and JWT secrets
   ```



### Local Development

#### Backend
1. `cd backend`
2. `cp .env.example .env` (Configure your Turso DB/Redis/SMTP)
3. `go mod download`
4. `go run cmd/api/main.go`

#### Frontend
1. `cd frontend`
2. `npm install`
3. `npm start`

## 🔑 Admin Setup

Registered users are regular users by default. To create an initial admin:


## 🌐 API Endpoints

### Public
- `POST /auth/register` - User registration
- `POST /auth/login` - User login (Sets HTTP-only cookie)
- `GET /scoreboard` - Get cached leaderboard
- `GET /notifications` - View active admin broadcasts

### Protected (User)
- `POST /challenges/:id/submit` - Submit flag (Rate limited)
- `POST /teams` - Create a team
- `POST /teams/join/:code` - Join a team via invite code

### Admin
- `POST /admin/challenges` - Create new challenge
- `POST /admin/notifications` - Broadcast an announcement
- `POST /admin/notifications/:id/toggle` - Activate/Deactivate broadcasts

## 📁 Project Structure

```
RootAccess/
├── backend/
│   ├── cmd/api/main.go          # API Entry point
│   ├── cmd/admin/main.go        # Admin CLI tool
│   ├── internal/
│   │   ├── database/            # Turso DB & Redis logic
│   │   ├── services/            # Business logic (Caching, Auth, etc.)
│   │   └── handlers/            # HTTP Controllers
├── frontend/
│   ├── src/app/components/      # Angular UI Components
│   └── src/app/services/        # Frontend API services
├── docker-compose.prod.yml      # Production orchestration
└── README.md
```

## 🛡️ Security Considerations

- **Secrets**: Never commit `.env` or `docker-compose.prod.yml` to version control.
- **JWT**: In production, ensure `JWT_SECRET` is a random 32+ character string.
- **SMTP**: Port 25 is often blocked by ISPs; use port 587 (STARTTLS) or 465 (SSL).

---

Made with ❤️ for the CTF community