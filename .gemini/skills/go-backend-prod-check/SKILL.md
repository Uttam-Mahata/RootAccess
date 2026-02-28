---
name: go-backend-prod-check
description: Audits Go backends for production readiness. Use when Gemini CLI needs to evaluate code for security, performance, logging, error handling, and production-standard configurations before deployment.
---

# Go Backend Production Readiness Auditor

This skill provides a procedural framework for auditing a Go-based backend (specifically targeting Gin/MongoDB/Redis stacks) to ensure it meets production-level engineering standards.

## Audit Workflow

### 1. Configuration & Secrets
- **Environment Variables**: Verify all sensitive data (DB URIs, API keys, JWT secrets) is loaded from environment variables via `godotenv` or similar. No hardcoded strings.
- **Default Values**: Check that defaults for critical security settings (like `JWT_SECRET`) are not used in production or trigger a failure if missing.
- **Trusted Proxies**: In production, ensure `r.SetTrustedProxies` is explicitly configured to prevent IP spoofing.

### 2. Security & Authentication
- **CORS Policy**: Ensure CORS is restricted to specific origins in production. Use `SameSite=Lax` or `Strict` for cookies.
- **JWT Hardening**: Check token expiration, signing algorithm validation, and secure storage (HTTP-only, Secure cookies).
- **Rate Limiting**: Verify that critical endpoints (Login, Register, Flag Submission) have IP-based or user-based rate limiting.
- **Input Validation**: Check that all request structs use binding tags (e.g., `binding:"required,email"`) and that the backend validates data types before processing.

### 3. Database Interactions (MongoDB/Redis)
- **Connection Pooling**: Verify that `MinPoolSize` and `MaxPoolSize` are configured for MongoDB.
- **Timeouts**: Ensure every DB operation uses a `context.Context` with a reasonable timeout (e.g., 5-10 seconds) to prevent hanging connections.
- **Indexing**: Identify frequently queried fields and ensure appropriate MongoDB indexes are created.
- **Caching**: Check if heavy read operations (Scoreboards, Leaderboards) are cached in Redis with sensible TTLs.

### 4. Error Handling & Logging
- **Error Propagation**: Ensure errors are wrapped with context and returned rather than ignored. Use `%w` for wrapping.
- **No Panics**: Verify that the code uses proper error returns instead of `panic()`. Ensure the Gin `Recovery()` middleware is used.
- **Logging**: Check that logs are structured (JSON preferred for production) and do not leak sensitive information (PII, tokens, hashes).
- **Client Messages**: Ensure internal error details are logged but NOT returned to the client (return generic "Internal Server Error" instead).

### 5. Performance & Concurrency
- **Goroutine Safety**: Check for proper use of `sync.Mutex` or `sync.RWMutex` when accessing shared state.
- **Resource Leaks**: Ensure `defer rows.Close()` or `defer cancel()` is called immediately after resource allocation.
- **Response Size**: Verify that large lists use pagination or projection to limit response size.

### 6. Observability & Docs
- **Swagger/OpenAPI**: Ensure all handlers have complete annotations and the `docs` are up to date.
- **Health Checks**: Verify the existence of a `/health` or `/ping` endpoint for load balancer/orchestrator monitoring.

## Detailed References

- **Security Details**: Read [references/security_checklist.md](references/security_checklist.md) for deep dives into JWT and CORS.
- **Idiomatic Go**: Read [references/go_patterns.md](references/go_patterns.md) for error handling and concurrency patterns.
