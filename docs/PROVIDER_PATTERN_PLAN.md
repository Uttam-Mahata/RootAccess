# Provider Pattern Implementation Plan - RootAccess

This document outlines the architectural shift to a Provider Pattern, enabling flexible selection of Database, Storage, and Cache technologies while maintaining "Always Free" deployment capabilities.

## 1. Objectives
- **Flexibility**: Swap between MongoDB (Document) and Turso (SQL) seamlessly.
- **Cost Optimization**: Enable AWS Lambda deployment without NAT Gateways by using public-facing but secure "Always Free" providers.
- **Resilience**: Maintain fallbacks for all services (e.g., Memory cache if Redis is unavailable).
- **Scalability**: Support horizontal scaling via distributed providers (Redis, S3).

## 2. Configuration (`Config` struct)
Update `backend/internal/config/config.go` to include:

| Variable | Possible Values | Description |
|----------|-----------------|-------------|
| `DB_TYPE` | `mongodb`, `turso` | Primary database technology. |
| `STORAGE_TYPE` | `s3`, `r2`, `local` | Object storage provider. |
| `CACHE_TYPE` | `redis`, `memory` | Caching and state sync provider. |
| `TURSO_URL` | string | libSQL connection URL. |
| `TURSO_TOKEN` | string | libSQL auth token. |
| `R2_ENDPOINT` | string | Cloudflare R2 S3-compatible endpoint. |
| `R2_ACCESS_KEY` | string | R2 Access Key. |
| `R2_SECRET_KEY` | string | R2 Secret Key. |

## 3. Database Layer (Repositories)
The current repositories are concrete implementations tied to MongoDB. We will move to an interface-driven approach.

### 3.1. Repository Interfaces
Define all repository methods in `backend/internal/repositories/interfaces/`:
- `UserRepository`
- `ChallengeRepository`
- `SubmissionRepository`
- `TeamRepository`
- ... and others.

### 3.2. Implementations
- **MongoDB**: Move existing logic from `internal/repositories/*.go` to `internal/repositories/mongodb/`.
- **Turso**: Create new SQL-based implementations in `internal/repositories/turso/` using `database/sql` and the `libsql` driver.

### 3.3. Factory Pattern
Create a `RepositoryFactory` in `internal/repositories/factory.go`:
```go
func NewUserRepository(cfg *config.Config) interfaces.UserRepository {
    if cfg.DB_TYPE == "turso" {
        return turso.NewUserRepository(database.TursoDB)
    }
    return mongodb.NewUserRepository(database.MongoDB)
}
```

## 4. Object Storage Layer
Create a new package `backend/internal/storage/`:

### 4.1. Interface
```go
type StorageProvider interface {
    UploadFile(ctx context.Context, bucket string, filename string, data []byte) error
    DeleteFile(ctx context.Context, bucket string, filename string) error
    GetDownloadURL(ctx context.Context, bucket string, filename string) (string, error)
}
```

### 4.2. Implementations
- **S3Provider**: Standard AWS S3 SDK.
- **R2Provider**: AWS S3 SDK with custom endpoint for Cloudflare.
- **LocalProvider**: File system storage (for development).

## 5. Cache Layer
Refine the existing logic to ensure consistent behavior across `redis` and `memory` modes.
- **Redis**: Use `go-redis`.
- **Memory**: Use `sync.Map` or a structured local cache with TTL.

## 6. Manual Backup & Purge Utility
Create `backend/cmd/storage-tool/main.go`:
- **Command `backup`**: Downloads all objects from a provider to local storage.
- **Command `verify`**: Compares checksums between local and remote storage.
- **Command `purge`**: Deletes objects older than N days from the remote provider.
- **Workflow**: `backup` -> `verify` -> `purge`.

## 7. Implementation Roadmap
1. **Infrastructure**: Add `DB_TYPE` and `STORAGE_TYPE` to config and `.env`.
2. **Interfaces**: Create the `interfaces` package and define all repository methods.
3. **Refactor MongoDB**: Move current code to the `mongodb` package.
4. **Initialize Turso**: Setup the libSQL driver and connection logic.
5. **Storage Scaffolding**: Create the `StorageProvider` interface and implement `LocalProvider`.
6. **Integration**: Update `routes.go` to use the `RepositoryFactory`.
7. **Turso/R2 Implementation**: Incrementally add Turso SQL queries and R2 client logic.

## 8. Data Model Considerations
Since MongoDB uses `ObjectID` and Turso will likely use `UUID` (string), we will:
- Standardize on `string` IDs in the `models` package.
- Use mapping logic in the repository layer to convert between `string` and `primitive.ObjectID`.

---
*Note: This plan maintains all existing MongoDB and Redis functionality while adding the infrastructure for Turso and R2.*
