# Database Migration: Add Format Fields

This migration adds `description_format` and `content_format` fields to existing challenges and writeups in the database.

## What it does

- Adds `description_format: "markdown"` to all challenges that don't have this field
- Adds `content_format: "markdown"` to all writeups that don't have this field
- Preserves all existing data
- Safe to run multiple times (idempotent)

## Prerequisites

- MongoDB must be running
- Go must be installed

## How to run

### Option 1: Using environment variables

```bash
cd scripts
export MONGO_URI="mongodb://localhost:27017"
export MONGO_DATABASE="ctf_platform"
go run add-format-fields.go
```

### Option 2: Using .env file

```bash
cd backend
source .env
cd ../scripts
go run add-format-fields.go
```

### Option 3: With Docker Compose

```bash
# Make sure containers are running
docker-compose up -d

# Run migration
docker-compose exec backend go run /app/scripts/add-format-fields.go
```

## Expected Output

```
Connecting to MongoDB: mongodb://localhost:27017
Database: ctf_platform

=== Migrating Challenges Collection ===
✓ Updated 5 challenges with description_format='markdown'

=== Migrating Writeups Collection ===
✓ Updated 3 writeups with content_format='markdown'

=== Migration Complete ===
Total challenges migrated: 5
Total writeups migrated: 3

All existing records now have format fields set to 'markdown' by default.
New records will require explicit format specification from the frontend.
```

## Verification

After running the migration, you can verify in MongoDB:

```javascript
// Check challenges
db.challenges.findOne({}, { description_format: 1 })

// Check writeups
db.writeups.findOne({}, { content_format: 1 })
```

Both should show the format field with value "markdown".
