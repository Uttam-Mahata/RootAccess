# RootAccess - AWS Deployment Guide

This document details the production deployment architecture for the RootAccess Backend on AWS Lambda and the Frontend on Vercel.

## 1. Backend Architecture (AWS)

The backend is deployed as a serverless application using the **AWS Serverless Application Model (SAM)**.

### Components:
- **AWS Lambda**: Runs the Go backend. Configured with 4GB RAM (~2.3 vCPUs) for optimal performance.
- **Amazon API Gateway**: Exposes the Lambda function as a REST API.
- **Amazon ElastiCache (Redis)**: Used for session management (OAuth state) and scoreboard caching.
  - **Node Type**: `cache.t3.micro`
  - **Replication**: 2 Nodes with Automatic Failover.
  - **Security**: Transit Encryption (TLS) and At-Rest Encryption enabled.
- **AWS Secrets Manager**: Centralized store for all environment variables and sensitive credentials.
- **Amazon VPC**:
  - **Private Subnets**: Houses Lambda and Redis for secure internal communication.
  - **Public Subnets**: Contains a NAT Gateway to allow Lambda to connect to the internet (required for MongoDB Atlas and OAuth providers).

### Configuration (Secrets Manager)
The backend loads all configuration from a JSON secret in Secrets Manager named `RootAccess/Backend/Secrets`. 

| Key | Description |
|-----|-------------|
| `MONGO_URI` | MongoDB Atlas connection string |
| `DB_NAME` | Database name |
| `JWT_SECRET` | Secret key for signing JWT tokens |
| `REDIS_ADDR` | Primary endpoint of ElastiCache |
| `REDIS_PASSWORD` | Redis Auth Token (min 16 chars) |
| `FRONTEND_URL` | URL of the deployed frontend (for CORS/Emails) |
| `GOOGLE_CLIENT_ID` | OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | OAuth Client Secret |
| `DISCORD_CLIENT_ID` | OAuth Client ID |
| `DISCORD_CLIENT_SECRET` | OAuth Client Secret |

---

## 2. Deployment Workflows

### Manual Deployment (Local)
Use the provided script in the `backend/` directory:
```bash
cd backend
./deploy_aws.sh
```
This script builds the Go binary, ensures secrets are prepared, and runs `sam deploy`.

### Automated Deployment (GitHub Actions)
Workflows are configured to trigger on pushes to the `main` branch.

#### Backend (AWS)
- **Path**: `.github/workflows/backend-deploy.yml`
- **IAM User**: `github-actions-rootaccess` (Dedicated user with restricted permissions)
- **Requires**: 
  - `AWS_ACCESS_KEY_ID`: Dedicated deployment key
  - `AWS_SECRET_ACCESS_KEY`: Dedicated deployment secret
  - `AWS_REGION`: us-east-1
  - `REDIS_AUTH_TOKEN`: `Daemon4rtacsRootAccess`

#### Frontend (Vercel)
- **Path**: `.github/workflows/frontend-deploy.yml`
- **Requires**:
  - `VERCEL_TOKEN`: Your Vercel personal access token.
  - `VERCEL_ORG_ID`: Found in your Vercel account settings.
  - `VERCEL_PROJECT_ID`: Found in your Vercel project settings.
  - `PROD_API_URL`: The backend API URL (e.g., `https://ctfapis.rootaccess.live`).

---

## 3. Technical Implementation Details

### Redis TLS Connection
AWS ElastiCache with transit encryption requires **TLS v1.2+**. The backend handles this automatically in `internal/database/redis.go`:
```go
if password != "" {
    options.TLSConfig = &tls.Config{
        MinVersion: tls.VersionTLS12,
    }
}
```

### Server Stability
To prevent `500 Internal Server Errors` during infrastructure maintenance, the `OAuthHandler` includes nil-checks for the Redis client. If Redis is unavailable, the system returns a `503 Service Unavailable` error instead of crashing.

### Networking Flow
1. User requests `ctfapis.rootaccess.live`.
2. **Cloudflare** CNAME points to **AWS API Gateway** (Edge).
3. **API Gateway** triggers **Lambda**.
4. **Lambda** (inside VPC) connects to:
   - **Redis** (locally within VPC).
   - **MongoDB Atlas** (outbound via NAT Gateway).
   - **OAuth Providers** (outbound via NAT Gateway).

---

## 4. Custom Domain Setup
1. **ACM Certificate**: Requested in `us-east-1` for `ctfapis.rootaccess.live`.
2. **DNS Validation**: CNAME record added to Cloudflare.
3. **API Mapping**: Custom domain created in API Gateway and mapped to the `Prod` stage.
4. **Final CNAME**: `ctfapis` in Cloudflare points to the AWS CloudFront distribution domain.
