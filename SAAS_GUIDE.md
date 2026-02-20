# RootAccess — SaaS / Multi-Tenant Guide

This guide explains the multi-tenant (SaaS) layer added to RootAccess and walks through
every step an event organizer needs to take — from registering an organization to writing
their first SDK-powered CTF frontend.

---

## Table of Contents

1. [What Problem Does This Solve?](#1-what-problem-does-this-solve)
2. [How It Works — The Big Picture](#2-how-it-works--the-big-picture)
3. [Concepts at a Glance](#3-concepts-at-a-glance)
4. [Step-by-Step: Register an Organization](#4-step-by-step-register-an-organization)
5. [Step-by-Step: Create a CTF Event](#5-step-by-step-create-a-ctf-event)
6. [Using the Event Token in an SDK Client](#6-using-the-event-token-in-an-sdk-client)
7. [Bring Your Own Database](#7-bring-your-own-database)
8. [Bring Your Own File Storage (S3)](#8-bring-your-own-file-storage-s3)
9. [Complete API Reference](#9-complete-api-reference)
10. [Security Notes](#10-security-notes)
11. [FAQ](#11-faq)

---

## 1. What Problem Does This Solve?

### Before: self-host everything

The original RootAccess model requires each organizer to:

1. Clone the repository
2. Configure MongoDB, Redis, SMTP on their own server
3. Build and deploy Docker containers
4. Maintain the infrastructure for the duration of the event

This is a significant operational burden, especially for a university club running its
first CTF or a company that wants to run an internal security competition.

### After: sign up and get an API key

With the SaaS layer, organizers:

1. **Register** at `rootaccess.live` — a single `POST /orgs` call
2. **Receive an API key** — one API call to create the org is all that's needed
3. **Create an event** — another single API call; receive an event token
4. **Point any SDK** at `rootaccess.live` with their event token
5. **Build whatever they want** on top — custom frontend, Discord bot, live overlay

The backend is already running. Organizers never touch infrastructure.

```
rootaccess.live (hosted Go backend)
        │
        ├── Org "Revelation"  ──► custom Angular frontend (TS client)
        ├── Org "NullCon"      ──► Discord bot (Python client)
        └── Org "ACME Corp"   ──► admin automation scripts (Go client)
```

---

## 2. How It Works — The Big Picture

```
┌─────────────────────────────────────────────────────────────┐
│               rootaccess.live  (this backend)               │
│                                                             │
│  POST /orgs          ←  Organizer registers an org          │
│  POST /orgs/:id/events ← Organizer creates a CTF event      │
│                                                             │
│  Protected routes using X-Event-Token header:               │
│    challenges, submissions, scoreboard, teams …             │
└─────────────────────────────────────────────────────────────┘
          ▲                           ▲
          │ X-API-Key                 │ X-Event-Token
          │                           │
   Organizer scripts          Participants' browser /
   (setup phase)              bot / custom frontend
```

**Two credential types:**

| Credential | Format | Used by | Used for |
|---|---|---|---|
| **API key** | `ra_org_<32 hex>` | Organizer only | Creating / managing events |
| **Event token** | `evt_<32 hex>` | SDK clients | Scoping all API calls to one event |

Both are shown **once** at creation time and never stored in plain-text — only bcrypt
hashes are persisted.

---

## 3. Concepts at a Glance

| Concept | Description |
|---|---|
| **Organization** | A team or company registered on the platform. Has a slug (e.g. `revelation`) and an API key. |
| **Event** | A single CTF competition run by an org. Has its own event token, start/end times, scoreboard settings, and optionally its own MongoDB URI and S3 bucket. |
| **API Key** | Long-lived secret for the org owner. Used to create/update events — keep it server-side. |
| **Event Token** | Short-lived-ish secret distributed to SDK clients. Scopes every API call to one event. |
| **Custom Mongo URI** | Optional. If supplied, participant data for that event lives in the org's own MongoDB, not the platform's shared database. |
| **S3 Config** | Optional. If supplied, challenge file uploads go to the org's own S3 bucket instead of the platform default. |

---

## 4. Step-by-Step: Register an Organization

### Using curl

```bash
curl -s -X POST https://rootaccess.live/orgs \
  -H "Content-Type: application/json" \
  -d '{
    "name":        "Revelation CTF Club",
    "owner_email": "admin@revelation.team",
    "owner_name":  "Alice"
  }'
```

**Response (201):**

```json
{
  "message": "Organization created successfully. Store your API key — it will not be shown again.",
  "org": {
    "id":           "665f1a2b3c4d5e6f7a8b9c0d",
    "name":         "Revelation CTF Club",
    "slug":         "revelation-ctf-club",
    "owner_email":  "admin@revelation.team",
    "owner_name":   "Alice",
    "api_key_prefix": "ra_org_abc12",
    "created_at":   "2025-07-01T10:00:00Z",
    "updated_at":   "2025-07-01T10:00:00Z"
  },
  "api_key": "ra_org_abc12def34567890abcdef1234567890"
}
```

> ⚠️ **Copy the `api_key` value and store it securely. It will not be shown again.**

### Custom slug

If you want a specific slug (used in future display URLs) pass it explicitly:

```bash
curl -s -X POST https://rootaccess.live/orgs \
  -H "Content-Type: application/json" \
  -d '{
    "name":        "Revelation CTF Club",
    "slug":        "revelation",
    "owner_email": "admin@revelation.team",
    "owner_name":  "Alice"
  }'
```

Slugs must match `^[a-z0-9-]+$`. If you leave `slug` blank, it is derived from `name`
automatically (spaces → hyphens, lowercase).

---

## 5. Step-by-Step: Create a CTF Event

Once you have an org ID and API key you can create as many events as you like.

### Minimal event (no custom DB, no S3)

```bash
ORG_ID="665f1a2b3c4d5e6f7a8b9c0d"
API_KEY="ra_org_abc12def34567890abcdef1234567890"

curl -s -X POST "https://rootaccess.live/orgs/${ORG_ID}/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "name":        "RevelationCTF 2025",
    "description": "Annual CTF hosted by the Revelation club.",
    "start_time":  "2025-09-01T09:00:00Z",
    "end_time":    "2025-09-03T09:00:00Z"
  }'
```

**Response (201):**

```json
{
  "message": "Event created successfully. Store your event token — it will not be shown again.",
  "event": {
    "id":                    "773a4b5c6d7e8f90a1b2c3d4",
    "org_id":                "665f1a2b3c4d5e6f7a8b9c0d",
    "name":                  "RevelationCTF 2025",
    "slug":                  "revelationctf-2025",
    "start_time":            "2025-09-01T09:00:00Z",
    "end_time":              "2025-09-03T09:00:00Z",
    "is_active":             false,
    "scoreboard_visibility": "public",
    "event_token_prefix":    "evt_ab12"
  },
  "event_token": "evt_ab12cd34ef567890abcdef1234567890"
}
```

> ⚠️ **Copy the `event_token` value and store it securely. It will not be shown again.**

### Full event (custom DB + S3 + scoreboard freeze)

```bash
curl -s -X POST "https://rootaccess.live/orgs/${ORG_ID}/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${API_KEY}" \
  -d '{
    "name":        "RevelationCTF 2025",
    "description": "Annual CTF. Participant data stays on our servers.",
    "start_time":  "2025-09-01T09:00:00Z",
    "end_time":    "2025-09-03T09:00:00Z",
    "freeze_time": "2025-09-03T07:00:00Z",
    "scoreboard_visibility": "public",
    "frontend_url": "https://ctf.revelation.team",

    "custom_mongo_uri": "mongodb+srv://ctfuser:secret@cluster0.abc.mongodb.net/revelation2025",

    "s3_config": {
      "endpoint":   "s3.amazonaws.com",
      "bucket":     "revelation-ctf-files",
      "region":     "us-east-1",
      "access_key": "AKIAIOSFODNN7EXAMPLE",
      "secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
      "public_url": "https://d1234567890.cloudfront.net"
    }
  }'
```

---

## 6. Using the Event Token in an SDK Client

The event token goes in the `X-Event-Token` request header. Every call is then
automatically scoped to that event — challenges, submissions, scoreboard, and teams all
operate in isolation from every other event on the platform.

### Python (rootaccessd-client)

```python
import rootaccessd_client
from rootaccessd_client.rest import ApiException

# Point at rootaccess.live, authenticate with your event token
configuration = rootaccessd_client.Configuration(
    host="https://rootaccess.live",
    api_key={"X-Event-Token": "evt_ab12cd34ef567890abcdef1234567890"}
)

with rootaccessd_client.ApiClient(configuration) as api_client:
    challenges_api = rootaccessd_client.ChallengesApi(api_client)

    try:
        challenges = challenges_api.get_challenges()
        for ch in challenges:
            print(f"{ch.title} [{ch.category}] — {ch.max_points} pts")
    except ApiException as e:
        print(f"Error: {e}")
```

### TypeScript / Node.js (@rootaccessd/client)

```typescript
import { Configuration, ChallengesApi } from "@rootaccessd/client";

const config = new Configuration({
  basePath: "https://rootaccess.live",
  apiKey: "evt_ab12cd34ef567890abcdef1234567890",  // sent as X-Event-Token
});

const api = new ChallengesApi(config);
const challenges = await api.getChallenges();
challenges.forEach(ch => console.log(ch.title, ch.maxPoints));
```

### curl (any shell script or Discord bot)

```bash
EVENT_TOKEN="evt_ab12cd34ef567890abcdef1234567890"

# Get challenges
curl -s https://rootaccess.live/challenges \
  -H "X-Event-Token: ${EVENT_TOKEN}"

# Submit a flag
curl -s -X POST "https://rootaccess.live/challenges/CHALLENGE_ID/submit" \
  -H "Content-Type: application/json" \
  -H "X-Event-Token: ${EVENT_TOKEN}" \
  -d '{"flag": "FLAG{example}"}'

# Get scoreboard
curl -s https://rootaccess.live/scoreboard \
  -H "X-Event-Token: ${EVENT_TOKEN}"
```

---

## 7. Bring Your Own Database

**Why?** Universities and companies with compliance requirements (FERPA, GDPR, ISO 27001)
may be unable to store participant data on third-party servers. By supplying a
`custom_mongo_uri` when creating an event, all data written during that event — user
accounts, submissions, team rosters, scores — goes into *your* MongoDB instance, not
the shared platform database.

**How it works:**

```
Without custom_mongo_uri          With custom_mongo_uri
─────────────────────────         ─────────────────────────────────────────
rootaccess.live/DB                Organization's own MongoDB (Atlas, on-prem…)
  └── revelationctf-2025            └── all participant data for that event
```

**What to supply:**

Any [MongoDB connection string](https://www.mongodb.com/docs/manual/reference/connection-string/)
that the backend can reach at runtime. Examples:

| Deployment | URI format |
|---|---|
| MongoDB Atlas | `mongodb+srv://user:pass@cluster0.abc.mongodb.net/dbname` |
| Self-hosted with auth | `mongodb://user:pass@192.0.2.10:27017/dbname` |
| Self-hosted without auth | `mongodb://192.0.2.10:27017/dbname` |

> **Security:** The `custom_mongo_uri` is stored encrypted in the platform's database and
> is **never** returned in API responses (`json:"-"`).

---

## 8. Bring Your Own File Storage (S3)

Challenge files (binaries, pcaps, images) can be large. By supplying an `s3_config` when
creating an event, all file uploads for that event are stored in your own S3-compatible
bucket — not the platform's default storage.

**Supported providers:** Any S3-compatible service:
- Amazon S3
- Cloudflare R2
- Backblaze B2
- MinIO (self-hosted)
- DigitalOcean Spaces
- Wasabi

**S3 config fields:**

| Field | Required | Description |
|---|---|---|
| `endpoint` | Yes | Storage endpoint, e.g. `s3.amazonaws.com` or `s3.us-west-002.backblazeb2.com` |
| `bucket` | Yes | Bucket name |
| `region` | Yes | Region string, e.g. `us-east-1` |
| `access_key` | Yes | IAM / account access key |
| `secret_key` | Yes | IAM / account secret key — **never returned by the API** |
| `public_url` | No | CDN or public bucket URL used to generate download links, e.g. `https://d123.cloudfront.net` |

**Minimum IAM policy (AWS) for the bucket:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:PutObject", "s3:GetObject", "s3:DeleteObject", "s3:ListBucket"],
      "Resource": [
        "arn:aws:s3:::revelation-ctf-files",
        "arn:aws:s3:::revelation-ctf-files/*"
      ]
    }
  ]
}
```

---

## 9. Complete API Reference

All org/event endpoints are public (no user login needed for the GET endpoints).
Write endpoints require `X-API-Key`.

### Organizations

#### `POST /orgs` — Register an organization

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Display name (2–100 chars) |
| `slug` | string | ❌ | URL-safe identifier; auto-derived from `name` if blank |
| `owner_email` | string | ✅ | Contact email for the org owner |
| `owner_name` | string | ✅ | Display name of the owner |

Returns `201` with the org object **and** the plain-text `api_key` (shown once only).

---

#### `GET /orgs/:id` — Get organization details

No auth required. Returns public org fields (no API key hash).

---

### Events

#### `POST /orgs/:id/events` — Create an event

Requires `X-API-Key` header matching the org's API key.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Event display name (2–100 chars) |
| `slug` | string | ❌ | Unique within the org; auto-derived from `name` |
| `description` | string | ❌ | Optional markdown description |
| `start_time` | RFC3339 | ✅ | When the contest opens |
| `end_time` | RFC3339 | ✅ | When the contest closes (`end_time > start_time`) |
| `freeze_time` | RFC3339 | ❌ | Scoreboard freezes at this time; submissions continue |
| `scoreboard_visibility` | string | ❌ | `"public"` (default), `"private"`, or `"hidden"` |
| `frontend_url` | string | ❌ | URL of the organizer's custom frontend |
| `custom_mongo_uri` | string | ❌ | Org-owned MongoDB URI (see [§7](#7-bring-your-own-database)) |
| `s3_config` | object | ❌ | Org-owned file storage (see [§8](#8-bring-your-own-file-storage-s3)) |

Returns `201` with the event object **and** the plain-text `event_token` (shown once only).

---

#### `GET /orgs/:id/events` — List events for an org

No auth required. Returns array of event objects.

---

#### `GET /events/:id` — Get a single event

No auth required. Returns the event object.

---

#### `PUT /events/:id` — Update an event

Requires `X-API-Key` header. Accepts the same fields as `POST /orgs/:id/events` plus:

| Field | Type | Description |
|---|---|---|
| `is_active` | bool | Activate or deactivate the contest |
| `is_paused` | bool | Pause or resume flag submissions |

---

## 10. Security Notes

- **API keys and event tokens are never stored in plain-text.** The backend persists only
  bcrypt hashes. Even if the database is compromised, the keys cannot be recovered.
- **Keys are single-use display.** After the `POST /orgs` or `POST /orgs/:id/events` response,
  the plain-text credential is gone. If you lose it, create a new event (token rotation
  is a natural workflow).
- **`custom_mongo_uri` is invisible to the API.** The field carries `json:"-"` throughout
  the Go codebase and is never included in any HTTP response.
- **`s3_config.secret_key` is invisible to the API.** Same treatment — `json:"-"`.
- **Rate limits still apply.** Flag submission is limited to 5 attempts per minute per
  challenge regardless of which event token is used.

---

## 11. FAQ

**Q: Do I need to self-host anything?**

No. The entire backend runs at `rootaccess.live`. You only need to write your custom
frontend or bot, point it at the API, and pass your event token.

---

**Q: Can one organization run multiple events?**

Yes. Call `POST /orgs/:id/events` multiple times. Each event gets its own independent
token, time window, and (optionally) its own MongoDB and S3 bucket.

---

**Q: What happens to my data after the event ends?**

If you supplied a `custom_mongo_uri`, your data lives entirely in your own database —
delete it at any time. If you used the platform default, contact the platform admin to
request an export or deletion.

---

**Q: I lost my event token. What do I do?**

Update the event via `PUT /events/:id` with your org API key to set a new
`frontend_url`, `s3_config`, or other fields, but note that **token rotation is not yet
supported** — if the token is lost you will need to create a new event.

---

**Q: My university's IT team requires all data to stay on our servers. Can I do that?**

Yes. Supply `custom_mongo_uri` pointing to your on-premises or cloud MongoDB instance
when creating the event. Participant data (users, submissions, scores, teams) for that
event will be stored exclusively in your database. The platform operator's shared
database only holds the event metadata itself (name, times, token hash).

---

**Q: The Python / TypeScript SDKs mention `X-Event-Token`. Where is that sent?**

The `EventTokenMiddleware` in the backend reads the `X-Event-Token` request header,
validates it against the bcrypt hash, and injects the resolved event into the Gin
context. Every downstream handler can then access `event`, `event_id`, and `org_id`
from the context, making the scope automatic — no per-handler code changes needed.
