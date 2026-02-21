# RootAccess - AWS Deployment Guide

This document details the production deployment architecture for the RootAccess Backend on AWS Lambda and the Frontend on Vercel.

## 1. Backend Architecture (AWS)

The backend is deployed as a serverless application using the **AWS Serverless Application Model (SAM)**.

### Components:
- **AWS Lambda**: Runs the Go backend. Configured with 512MB RAM for optimal balance of performance and free tier sustainability.
- **Amazon API Gateway (REST)**: Exposes the standard API endpoints.
- **Amazon API Gateway (WebSocket)**: Handles persistent WebSocket connections.
- **Amazon ElastiCache (Redis)**: Used for session management, scoreboard caching, and tracking WebSocket Connection IDs.
  - **Node Type**: `cache.t3.micro`
  - **Replication**: 2 Nodes with Automatic Failover.
  - **Security**: Transit Encryption (TLS) and At-Rest Encryption enabled.
- **AWS Secrets Manager**: Centralized store for all environment variables and sensitive credentials.
- **Amazon VPC**:
  - **Private Subnets**: Houses Lambda and Redis for secure internal communication.
  - **Public Subnets**: Contains a NAT Gateway to allow Lambda to connect to the internet.

### Configuration (Secrets Manager)
The backend loads all configuration from a JSON secret in Secrets Manager named `RootAccess/Backend/Secrets`. 

| Key | Description |
|-----|-------------|
| `MONGO_URI` | MongoDB Atlas connection string |
| `DB_NAME` | Database name |
| `JWT_SECRET` | Secret key for signing JWT tokens |
| `REDIS_ADDR` | Primary endpoint of ElastiCache |
| `REDIS_PASSWORD` | Redis Auth Token (min 16 chars) |
| `FRONTEND_URL` | URL of the deployed frontend |
| `WS_CALLBACK_URL`| URL used by Lambda to push messages to WebSocket clients |
| `GOOGLE_CLIENT_ID`| OAuth Client ID |
| `GOOGLE_CLIENT_SECRET`| OAuth Client Secret |

---

## 2. WebSocket Architecture (Stateless)

Since AWS Lambda is short-lived, it cannot maintain persistent WebSocket connections. RootAccess uses a stateless architecture:

1.  **Connection Management**: API Gateway manages the persistent TCP connection with the client.
2.  **Tracking**: Upon connection (`$connect`), Lambda extracts the `user_id` from the JWT `token` query parameter and stores the `ConnectionID` in **Redis**.
3.  **Pushing Notifications**: When a broadcast event occurs (e.g., a solve), Lambda fetches all active `ConnectionIDs` from Redis and uses the `WS_CALLBACK_URL` via the AWS SDK (`apigatewaymanagementapi`) to push the message to each client.
4.  **Cleanup**: When a user disconnects, the `$disconnect` event triggers a Lambda to remove the ID from Redis.

---

## 3. Deployment Workflows

### Manual Deployment (Local)
Use the provided script in the `backend/` directory:
```bash
cd backend
./deploy_aws.sh
```
This script builds the Go binary, ensures secrets are prepared, and runs `sam deploy`. It also automatically syncs the `WS_CALLBACK_URL` from the stack outputs to Secrets Manager.

### Automated Deployment (GitHub Actions)
Workflows are configured to trigger on pushes to the `main` branch.

#### Backend (AWS)
- **Path**: `.github/workflows/backend-deploy.yml`
- **Requires**: 
  - `AWS_ACCESS_KEY_ID`: Dedicated deployment key
  - `AWS_SECRET_ACCESS_KEY`: Dedicated deployment secret
  - `REDIS_AUTH_TOKEN`: Your rotated Redis password.

#### Frontend (Vercel)
- **Path**: `.github/workflows/frontend-deploy.yml`
- **Requires**:
  - `PROD_API_URL`: The backend API URL (e.g., `https://ctfapis.rootaccess.live`).
  - `PROD_WS_URL`: The WebSocket WSS URL (e.g., `wss://gjbz5ce1bf.execute-api.us-east-1.amazonaws.com/Prod`).

---

## 4. Technical Implementation Details

### WebSocket Authentication
Standard cookies are not automatically sent with WebSocket requests across different domains. RootAccess authenticates WebSocket connections by appending the JWT token to the URL:
`wss://<api-id>.execute-api.us-east-1.amazonaws.com/Prod?token=<JWT_HERE>`

The backend validates this token during the `$connect` handshake.

### Redis TLS Connection
AWS ElastiCache with transit encryption requires **TLS v1.2+**. The backend handles this automatically in `internal/database/redis.go`.
