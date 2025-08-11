# Storage Design

This document describes the data storage architecture and database schema for Michishirube.

## Technology Choice

**SQLite** is used as the primary database for the following reasons:

- **Single User**: Application is designed for personal use
- **Simplicity**: No separate database server required
- **Portability**: Database is a single file that can be backed up easily
- **Performance**: Excellent for read-heavy workloads with moderate data volumes
- **Zero Configuration**: No setup or maintenance required

## Database Schema

### Tables

#### tasks
Main table storing task information.

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    jira_id TEXT NOT NULL,
    title TEXT NOT NULL,
    priority TEXT NOT NULL CHECK(priority IN ('minor', 'normal', 'high', 'critical')),
    status TEXT NOT NULL CHECK(status IN ('new', 'in_progress', 'blocked', 'done', 'archived')),
    tags TEXT, -- JSON array: ["frontend", "bug", "urgent"]
    blockers TEXT, -- JSON array: ["waiting for review", "blocked by OCPBUGS-5678"]
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### links
Related resources for tasks (PRs, Slack threads, etc.).

```sql
CREATE TABLE links (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('pull_request', 'slack_thread', 'jira_ticket', 'documentation', 'other')),
    url TEXT NOT NULL,
    title TEXT,
    status TEXT, -- "merged", "open", "resolved", etc.
    metadata TEXT, -- JSON for type-specific data
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

#### comments
User notes and comments on tasks.

```sql
CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

#### schema_migrations
Tracks applied database migrations.

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

Performance-critical indexes for common query patterns:

```sql
-- Task queries
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_jira_id ON tasks(jira_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_priority ON tasks(priority);

-- Link queries
CREATE INDEX idx_links_task_id ON links(task_id);
CREATE INDEX idx_links_type ON links(type);

-- Comment queries
CREATE INDEX idx_comments_task_id ON comments(task_id);

-- Search optimization
CREATE INDEX idx_tasks_title ON tasks(title);
```

## Storage Interface

The storage layer is abstracted through interfaces to allow for future extensibility:

```go
type Storage interface {
    // Tasks
    CreateTask(task *models.Task) error
    GetTask(id string) (*models.Task, error)
    ListTasks(filters TaskFilters) ([]*models.Task, error)
    UpdateTask(task *models.Task) error
    DeleteTask(id string) error
    
    // Links
    CreateLink(link *models.Link) error
    GetLinksForTask(taskID string) ([]*models.Link, error)
    UpdateLink(link *models.Link) error
    DeleteLink(id string) error
    
    // Comments
    CreateComment(comment *models.Comment) error
    GetCommentsForTask(taskID string) ([]*models.Comment, error)
    DeleteComment(id string) error
    
    // Search
    SearchTasks(query string, includeArchived bool) ([]*models.Task, error)
    
    // Maintenance
    Close() error
}
```

## Query Patterns

### Common Queries

**Active Tasks Dashboard:**
```sql
SELECT * FROM tasks 
WHERE status IN ('new', 'in_progress', 'blocked') 
ORDER BY priority DESC, created_at DESC;
```

**Search Tasks:**
```sql
SELECT * FROM tasks 
WHERE (title LIKE ? OR jira_id LIKE ? OR tags LIKE ?)
  AND (? = true OR status != 'archived')
ORDER BY created_at DESC;
```

**Task with Links and Comments:**
```sql
-- Task
SELECT * FROM tasks WHERE id = ?;

-- Links
SELECT * FROM links WHERE task_id = ? ORDER BY type;

-- Comments  
SELECT * FROM comments WHERE task_id = ? ORDER BY created_at;
```

### Performance Considerations

1. **Pagination**: Implement LIMIT/OFFSET for large result sets
2. **JSON Fields**: Use JSON functions for complex tag/blocker queries
3. **Full-Text Search**: Consider FTS5 extension for advanced search features
4. **Connection Pooling**: Single connection for SQLite (no concurrency issues)

## Data Integrity

### Constraints
- **Foreign Keys**: Ensure referential integrity with CASCADE DELETE
- **Check Constraints**: Validate enum values at database level
- **NOT NULL**: Critical fields must have values

### Backup Strategy
- **File-based**: Simple file copy of the SQLite database
- **Export**: JSON export functionality for data portability
- **Incremental**: WAL mode for better concurrency and easier backups

## Migration Strategy

See [migrations.md](./migrations.md) for detailed migration system design.