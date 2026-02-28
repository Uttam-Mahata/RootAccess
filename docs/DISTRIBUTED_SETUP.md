# Distributed Cluster Setup

This guide explains how to deploy the RootAccess CTF Platform in a distributed environment (multi-node cluster) for high availability and performance.

## 🏗️ Architecture Overview

![Distributed Architecture Diagram](./arch_dist.png)

A standard distributed setup consists of:
1.  **Gateway (1 Node):** Nginx acting as a Load Balancer and Reverse Proxy.
2.  **Application Nodes (3+ Nodes):** Instances of the Go backend and Angular frontend.
3.  **Database & Cache (1-3 Nodes):** MongoDB for persistence and Redis for session/scoreboard caching.

### How Communication Flows

*   **Gateway (Nginx) → App Nodes:**
    *   **Protocol:** HTTP/HTTPS.
    *   **Mechanism:** Nginx "reverse proxies" traffic to the IP addresses of your App VMs on the configured port (default: 8080).
    *   **Security:** Relies on internal network firewall rules. SSH keys are **not** required for this communication.
*   **App Nodes → Database (MongoDB):**
    *   **Protocol:** TCP (MongoDB Wire Protocol).
    *   **Mechanism:** The Go backend connects using a Connection String (e.g., `mongodb://user:password@db_ip:27017/ctfd`).
    *   **Security:** Relies on MongoDB authentication and IP whitelisting.
*   **App Nodes → Cache (Redis):**
    *   **Protocol:** RESP (Redis Serialization Protocol).
    *   **Mechanism:** All nodes connect to a central Redis instance for shared scoreboard state and session management.

---

## 🐳 Deployment via Docker (Recommended)

To ensure your applications **keep running** automatically after crashes or reboots, we recommend running them as Docker containers with a restart policy.

### 1. Backend Service (App Nodes)
You can build the backend image on the node itself or push it to a registry.

**Dockerfile:** Provided in `backend/Dockerfile`.

```bash
# On each App Node:
cd backend
docker build -t ctf-backend .

# Run with auto-restart
docker run -d \
  --name ctf-backend \
  --restart always \
  -p 8080:8080 \
  -e MONGO_URI="mongodb://db_user:pass@db_node_ip:27017/ctf" \
  -e REDIS_ADDR="db_node_ip:6379" \
  -e JWT_SECRET="your_shared_secret_key" \
  ctf-backend
```

### 2. Frontend Service (App Nodes)
The provided `frontend/Dockerfile` is designed to serve **pre-built** artifacts. This is ideal for all architectures (x64, arm64, s390x) as it avoids building on the server.

**Step A: Build Locally (Workstation)**
```bash
cd frontend
npm install
npm run build --prod
# This creates dist/frontend/browser/
```

**Step B: Package & Run (App Node)**
Transfer the `frontend/` directory (including the new `dist/`) to your server, then:

```bash
cd frontend
docker build -t ctf-frontend .

# Run Nginx container
docker run -d \
  --name ctf-frontend \
  --restart always \
  -p 80:80 \
  ctf-frontend
```

### 3. Database & Cache (Data Node)
On your dedicated Database VM, you can use Docker Compose to manage MongoDB and Redis.

```yaml
# docker-compose.db.yml
version: '3.8'
services:
  mongodb:
    image: mongo:7.0
    restart: always
    ports: ["27017:27017"]
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: secure_password
    volumes:
      - mongo_data:/data/db

  redis:
    image: redis:7-alpine
    restart: always
    ports: ["6379:6379"]
    command: redis-server --requirepass secure_redis_password

volumes:
  mongo_data:
```

---

## 🔑 Critical: What You MUST Share

In a distributed setup, consistency across nodes is vital. If the following are not shared, users will experience session drops, 404 errors on files, or inconsistent scoring.

### 1. Shared `JWT_SECRET`
All backend instances must have the **exact same** `JWT_SECRET` in their `.env` files (or environment variables).

### 2. Shared Storage (Uploads)
If a user uploads a file to Node A, Node B must be able to see it.
*   **Docker Solution:** Mount a shared NFS volume to the same path in your containers.
    ```bash
    docker run -v /mnt/shared_nfs/uploads:/app/uploads ...
    ```

### 3. Centralized Redis
All application nodes must point to the **same Redis instance** (defined in `REDIS_ADDR`).

---

## ⚠️ Platform Limitations (s390x / IBM Z)

If you are deploying on **s390x architecture**, the frontend build tools (specifically `lightningcss` in Tailwind CSS v4) may not run natively.

**Solution:**
Use the **Docker Frontend** method described above.
1.  **Build** the Angular app on your local machine (x64).
2.  **Copy** the `frontend` folder (with `dist/`) to the s390x server.
3.  **Run** the Docker build command on the server. Since the `Dockerfile` only uses the `nginx:alpine` image (which supports s390x) and copies static files, it will work perfectly.

---

## 🛠️ Nginx Gateway Configuration

Configure your Gateway VM to load balance traffic to your App Nodes:

```nginx
upstream ctf_backend {
    least_conn;
    server 10.0.0.10:8080;
    server 10.0.0.11:8080;
    server 10.0.0.12:8080;
}

upstream ctf_frontend {
    least_conn;
    server 10.0.0.10:80;
    server 10.0.0.11:80;
    server 10.0.0.12:80;
}

server {
    listen 80;
    server_name ctf.yourdomain.com;

    # Proxy API requests to Backend Cluster
    location /api/ {
        proxy_pass http://ctf_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Proxy other requests to Frontend Cluster
    location / {
        proxy_pass http://ctf_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
