# Security Deep Dive Checklist

## JWT Hardening
- **Algorithm**: Always explicitly check the signing method in the parser (e.g., `*jwt.SigningMethodHMAC`).
- **Expiry**: Tokens should have a reasonable TTL (e.g., 15 minutes for access tokens, longer for refresh tokens).
- **Storage**: Prefer HTTP-only, Secure, SameSite=Lax cookies over localStorage.
- **Payload**: Never store sensitive data like passwords or internal IDs in the JWT payload.

## CORS Configuration
- **Origins**: Avoid `Allow-All (*)` in production. Use a whitelist loaded from configuration.
- **Methods**: Restrict allowed HTTP methods to only those actually used.
- **Credentials**: If using cookies, `AllowCredentials` must be `true` and the origin cannot be `*`.

## Rate Limiting
- **Authentication**: Limit attempts on `/auth/login` and `/auth/register` by IP.
- **Business Logic**: Limit actions like flag submission or team creation to prevent spam/automated attacks.
- **Headers**: Return standard `X-RateLimit-*` headers to inform clients.

## Input Sanitization
- **Shell Commands**: Avoid `exec.Command` with user-provided input. If necessary, use strict whitelisting.
- **NoSQL Injection**: Even with drivers like `mongo-driver`, ensure user input isn't used to build raw BSON queries without proper typing.
