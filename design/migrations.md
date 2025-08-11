# Database Migrations

This document describes the automatic database migration system for Michishirube.

## Overview

The migration system ensures that the database schema evolves smoothly as the application develops, without requiring manual intervention from users. All migrations are applied automatically when the application starts.

## Design Principles

### Transparent to Users
- Migrations run automatically on application startup
- No user intervention required
- Backward compatibility maintained where possible
- Clear logging of migration activities

### Safe and Reliable
- Migrations are atomic (all-or-nothing)
- Database backups before major schema changes
- Rollback capability for critical failures
- Version tracking to prevent duplicate applications

### Development Friendly
- Easy to add new migrations during development
- Clear naming conventions for migration files
- Support for both schema changes and data transformations

## Migration System Architecture

### Schema Migrations Table

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    checksum TEXT -- Optional: for integrity verification
);
```

### Migration File Structure

```
internal/storage/sqlite/migrations/
├── 001_initial_schema.sql
├── 002_add_priority_field.sql
├── 003_create_links_table.sql
├── 004_add_status_blocked.sql
└── 005_add_indexes.sql
```

### Migration File Format

Each migration file contains:

1. **Header Comment**: Description and metadata
2. **Up Migration**: Schema changes to apply
3. **Down Migration**: (Optional) Rollback instructions

**Example Migration File (`002_add_priority_field.sql`):**

```sql
-- Migration: Add priority field to tasks
-- Version: 002
-- Description: Add priority field with default value
-- Author: System
-- Date: 2024-01-15

-- UP MIGRATION
ALTER TABLE tasks ADD COLUMN priority TEXT DEFAULT 'normal' 
CHECK(priority IN ('minor', 'normal', 'high', 'critical'));

-- Update existing rows to have default priority
UPDATE tasks SET priority = 'normal' WHERE priority IS NULL;

-- Make priority field NOT NULL after setting defaults
-- Note: SQLite doesn't support ALTER COLUMN, so we use a workaround if needed

-- DOWN MIGRATION (commented for reference)
-- ALTER TABLE tasks DROP COLUMN priority;
```

## Migration Process

### Startup Flow

1. **Initialize Migration Table**: Create `schema_migrations` if it doesn't exist
2. **Scan Migration Files**: Read all migration files from the migrations directory
3. **Compare Versions**: Determine which migrations need to be applied
4. **Backup Database**: Create backup before applying migrations (optional)
5. **Apply Migrations**: Execute migrations in order within transactions
6. **Update Tracking**: Record successful migrations in the tracking table
7. **Log Results**: Report migration status to application logs

### Error Handling

```go
type MigrationError struct {
    Version int
    Name    string
    Error   error
    Context string
}

func (e *MigrationError) Error() string {
    return fmt.Sprintf("migration %d (%s) failed: %v", e.Version, e.Name, e.Error)
}
```

**Error Scenarios:**
- **Invalid SQL**: Log error and stop application startup
- **Constraint Violations**: Rollback transaction and stop
- **Missing Files**: Warning for gaps in version sequence
- **Checksum Mismatch**: Detect if migration files were modified

## Implementation

### Migration Manager Interface

```go
type MigrationManager interface {
    // Apply all pending migrations
    ApplyMigrations() error
    
    // Get current schema version
    GetCurrentVersion() (int, error)
    
    // List pending migrations
    GetPendingMigrations() ([]Migration, error)
    
    // Force apply specific migration (development only)
    ApplyMigration(version int) error
}

type Migration struct {
    Version     int
    Name        string
    Filename    string
    SQL         string
    Checksum    string
    AppliedAt   *time.Time
}
```

### Migration Execution

```go
func (m *SQLiteMigrationManager) ApplyMigrations() error {
    pending, err := m.GetPendingMigrations()
    if err != nil {
        return err
    }
    
    if len(pending) == 0 {
        log.Println("No pending migrations")
        return nil
    }
    
    log.Printf("Applying %d migrations", len(pending))
    
    for _, migration := range pending {
        if err := m.applyMigration(migration); err != nil {
            return &MigrationError{
                Version: migration.Version,
                Name:    migration.Name,
                Error:   err,
                Context: "applying migration",
            }
        }
        log.Printf("Applied migration %d: %s", migration.Version, migration.Name)
    }
    
    return nil
}

func (m *SQLiteMigrationManager) applyMigration(migration Migration) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Execute migration SQL
    if _, err := tx.Exec(migration.SQL); err != nil {
        return err
    }
    
    // Record migration as applied
    _, err = tx.Exec(`
        INSERT INTO schema_migrations (version, name, checksum) 
        VALUES (?, ?, ?)
    `, migration.Version, migration.Name, migration.Checksum)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Migration Examples

### Initial Schema (001_initial_schema.sql)

```sql
-- Migration: Initial database schema
-- Version: 001
-- Description: Create initial tables for tasks, links, and comments

CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    jira_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('new', 'in_progress', 'done', 'archived')),
    tags TEXT,
    blockers TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE links (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('pull_request', 'slack_thread', 'jira_ticket', 'documentation', 'other')),
    url TEXT NOT NULL,
    title TEXT,
    status TEXT,
    metadata TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Initial indexes
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_jira_id ON tasks(jira_id);
CREATE INDEX idx_links_task_id ON links(task_id);
CREATE INDEX idx_comments_task_id ON comments(task_id);
```

### Adding New Status (004_add_status_blocked.sql)

```sql
-- Migration: Add blocked status
-- Version: 004
-- Description: Add 'blocked' status to task status enum

-- SQLite doesn't support ALTER CHECK constraints directly
-- We need to recreate the table with the new constraint

-- Create new table with updated constraint
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,
    jira_id TEXT NOT NULL,
    title TEXT NOT NULL,
    priority TEXT NOT NULL CHECK(priority IN ('minor', 'normal', 'high', 'critical')),
    status TEXT NOT NULL CHECK(status IN ('new', 'in_progress', 'blocked', 'done', 'archived')),
    tags TEXT,
    blockers TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table
INSERT INTO tasks_new SELECT * FROM tasks;

-- Drop old table and rename new one
DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

-- Recreate indexes
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_jira_id ON tasks(jira_id);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

## Development Workflow

### Adding New Migrations

1. **Create Migration File**: Use next version number
2. **Write Migration SQL**: Include both schema and data changes
3. **Test Migration**: Verify on development database
4. **Add to Version Control**: Commit migration file with code changes

### Testing Migrations

```bash
# Apply migrations to test database
go run cmd/server/main.go --migrate-only --db-path=test.db

# Verify schema
sqlite3 test.db ".schema"

# Check migration status
sqlite3 test.db "SELECT * FROM schema_migrations;"
```

## Best Practices

### Migration Guidelines

1. **Always Backward Compatible**: New migrations should not break existing data
2. **Atomic Operations**: Each migration should be a single atomic change
3. **Default Values**: Provide sensible defaults for new NOT NULL columns
4. **Data Preservation**: Never lose existing user data
5. **Test Thoroughly**: Test migrations on copies of production data

### Naming Conventions

- `{version}_{descriptive_name}.sql`
- Use sequential version numbers (001, 002, 003...)
- Use underscores and lowercase for names
- Be descriptive but concise

### Performance Considerations

- **Large Tables**: Consider impact of schema changes on large datasets
- **Indexes**: Add indexes in separate migrations for better visibility
- **Batching**: For data migrations, consider processing in batches
- **Timing**: Apply migrations during maintenance windows if needed