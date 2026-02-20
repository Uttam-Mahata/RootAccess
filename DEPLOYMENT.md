# RootAccess — Self-Hosting Deployment Guide

This guide walks you through deploying RootAccess for a live CTF event, from a fresh
clone to a running platform. It covers a college or organization hosting an internal or
public competition.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Clone & Configure](#2-clone--configure)
3. [Configure the Frontend](#3-configure-the-frontend)
4. [Build & Run](#4-build--run)
5. [Create the First Admin](#5-create-the-first-admin)
6. [Set Up Your Contest](#6-set-up-your-contest)
7. [Registration Modes](#7-registration-modes)
8. [Running the Event](#8-running-the-event)
9. [After the Event](#9-after-the-event)

---

## 1. Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Docker | 24+ | Required for containerized deployment |
| Docker Compose | v2 (plugin) | Bundled with Docker Desktop; `docker compose` (no hyphen) |
| Domain name or static IP | — | Used in frontend config and OAuth redirect URLs |
| SMTP credentials | — | For email verification and password resets |
| (Optional) OAuth app credentials | — | Google / GitHub / Discord — skip if not needed |

**Check Docker is ready:**
```bash
docker compose version   # should print Compose version v2.x.x
```

---

## 2. Clone & Configure

```bash
git clone https://github.com/Uttam-Mahata/RootAccess.git
cd RootAccess
cp backend/.env.example backend/.env
```

Open `backend/.env` and fill in each variable:

### Required Variables

| Variable | What to set |
|---|---|
| `JWT_SECRET` | Run `openssl rand -base64 32` and paste the output. **Never use the placeholder.** |
| `MONGO_URI` | `mongodb://mongo:27017` when using the prod compose file (service name = `mongo`) |
| `DB_NAME` | Any name, e.g. `ctf_event` |
| `FRONTEND_URL` | Your frontend URL, e.g. `https://ctf.college.edu` |

### SMTP (Email Verification)

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=yourbot@gmail.com
SMTP_PASS=xxxx xxxx xxxx xxxx   # Gmail App Password (not your login password)
SMTP_FROM=RootAccess CTF <yourbot@gmail.com>
```

> **Gmail App Password:** Go to your Google account → Security → 2-Step Verification →
> App passwords. Generate one for "Mail".

### Redis

```bash
REDIS_ADDR=redis:6379    # use the service name from compose
REDIS_PASSWORD=          # leave blank if unauthenticated
REDIS_DB=0
```

### Registration Access Control

```bash
# Options: open | domain | disabled
REGISTRATION_MODE=open

# Only used when REGISTRATION_MODE=domain
REGISTRATION_ALLOWED_DOMAINS=college.edu
```

See [Section 7](#7-registration-modes) for a full explanation.

### OAuth (Optional)

Skip any provider you don't want. Leave the variables blank and users won't see that
login option.

**Google:**
1. Open [Google Cloud Console](https://console.cloud.google.com)
2. Create a project → APIs & Services → Credentials → OAuth 2.0 Client ID (Web application)
3. Add Authorized redirect URI: `https://your-backend-host/auth/google/callback`

```bash
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=https://your-backend-host/auth/google/callback
```

**GitHub:**
1. GitHub → Settings → Developer settings → OAuth Apps → New OAuth App
2. Authorization callback URL: `https://your-backend-host/auth/github/callback`

```bash
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...
GITHUB_REDIRECT_URL=https://your-backend-host/auth/github/callback
```

**Discord:**
1. [Discord Developer Portal](https://discord.com/developers/applications) → New Application → OAuth2
2. Add redirect: `https://your-backend-host/auth/discord/callback`

```bash
DISCORD_CLIENT_ID=...
DISCORD_CLIENT_SECRET=...
DISCORD_REDIRECT_URL=https://your-backend-host/auth/discord/callback
```

---

## 3. Configure the Frontend

Edit `frontend/src/environments/environment.prod.ts` to point at your backend host:

```typescript
export const environment = {
  production: true,
  apiUrl: 'https://your-backend-host',        // or http://IP:8080
  wsUrl: 'wss://your-backend-host',           // or ws://IP:8080
  googleAuthUrl: 'https://your-backend-host/auth/google'
};
```

Replace `your-backend-host` with the actual domain or IP:port where the backend will be
reachable from users' browsers. If you are not using Google OAuth, the `googleAuthUrl`
value doesn't matter.

---

## 4. Build & Run

The production compose file builds both the backend and frontend images locally:

```bash
docker compose -f docker-compose.prod.example.yml up -d --build
```

**Port mapping:**

| Container | Host port | Service |
|---|---|---|
| `go_ctf_backend` | 8080 | Go/Gin API |
| `go_ctf_frontend` | 4200 → 80 | Angular (served by nginx) |

> If you are running behind a reverse proxy (nginx, Caddy, Traefik), proxy the public
> domain to port 8080 (backend) and port 4200 (frontend), and set `TRUSTED_PROXIES` to
> the proxy's internal IP so that client IPs are recorded correctly.

**Verify services are up:**
```bash
docker compose -f docker-compose.prod.example.yml ps
curl http://localhost:8080/health   # should return {"status":"ok"}
```

**View logs:**
```bash
docker compose -f docker-compose.prod.example.yml logs -f backend
```

---

## 5. Create the First Admin

The default registration role is `user`. Promote a user to admin using the interactive
CLI tool that runs inside the backend container:

```bash
docker exec -it go_ctf_backend ./admin-tool
```

Follow the prompts to create a new admin account or promote an existing user. This is
the only way to create admin accounts — the API does not expose a role-promotion endpoint.

---

## 6. Set Up Your Contest

Log in to the frontend with your admin account. The admin dashboard is available at
`/admin`.

**Checklist before opening registration:**

- [ ] **Create challenges** — Upload files, write descriptions, set categories and point
      values. Dynamic scoring (CTFd formula) adjusts points automatically as solves
      accumulate.
- [ ] **Set contest times** — Configure start time, end time, and (optionally) scoreboard
      freeze time. The scoreboard freezes at the freeze time but submissions continue
      until end time.
- [ ] **Configure registration mode** — Set `REGISTRATION_MODE` in your `.env` and
      restart the backend (see [Section 7](#7-registration-modes)).
- [ ] **Test a flag submission** — Create a test team/user and verify end-to-end.
- [ ] **Send the URL to participants** — Point them to the frontend URL.

---

## 7. Registration Modes

Set `REGISTRATION_MODE` in `backend/.env` and restart the backend container for the
change to take effect.

| Mode | `REGISTRATION_MODE` | `REGISTRATION_ALLOWED_DOMAINS` | Use case |
|---|---|---|---|
| **Open** | `open` (default) | ignored | Public CTF, multi-college event |
| **Domain-restricted** | `domain` | e.g. `college.edu` | Internal college-only CTF |
| **Disabled** | `disabled` | ignored | Invite-only; admin creates accounts via CLI |

### Domain-restricted example

Only `@college.edu` and `@cs.college.edu` addresses may register:

```bash
REGISTRATION_MODE=domain
REGISTRATION_ALLOWED_DOMAINS=college.edu,cs.college.edu
```

- Spaces around commas are ignored.
- The check uses suffix matching on the full email: `alice@college.edu` passes,
  `alice@notcollege.edu` does not.
- If `REGISTRATION_ALLOWED_DOMAINS` is left blank while mode is `domain`, no one can
  register (fails closed — safe default).

### Disabled mode

```bash
REGISTRATION_MODE=disabled
```

Anyone who tries to register receives: `"registration is currently closed"`. Admins can
still create accounts via the CLI tool (see [Section 5](#5-create-the-first-admin)).

### Applying a change mid-event

```bash
# Edit .env, then restart the backend container only:
docker compose -f docker-compose.prod.example.yml restart backend
```

---

## 8. Running the Event

- **Pause/resume:** Set start time in the future to delay the contest start; extend end
  time to allow more time.
- **Broadcast notifications:** Use the admin dashboard to send announcements to all
  connected participants (delivered via WebSocket).
- **Scoreboard freeze:** Once the freeze time is reached, the public scoreboard stops
  updating (but submissions continue). Unfreeze from the admin panel after the event ends
  to reveal the final standings.
- **First blood:** The WebSocket hub broadcasts first-blood events automatically when a
  challenge is solved for the first time.
- **Rate limiting:** Flag submission is rate-limited to 5 attempts per minute per
  challenge per user. No configuration needed.

---

## 9. After the Event

- **View final standings:** Unfreeze the scoreboard from the admin dashboard to reveal
  final rankings to all participants.
- **Export results:** Download the scoreboard and solve data from the admin dashboard for
  archiving or award ceremonies.
- **View analytics:** The admin panel shows per-challenge solve counts and score
  distribution over time.
- **Archive:** Take a MongoDB dump before shutting down:
  ```bash
  docker exec go_ctf_backend mongodump --uri="$MONGO_URI" --out=/tmp/dump
  docker cp go_ctf_backend:/tmp/dump ./ctf-dump-$(date +%Y%m%d)
  ```
- **Shutdown:**
  ```bash
  docker compose -f docker-compose.prod.example.yml down
  ```
  Add `-v` to also remove volumes (deletes all data — do this only after archiving).
