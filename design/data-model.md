# Data Model

This document describes the core data structures used in Michishirube.

## Core Entities

### Task

The main entity representing a work item or task.

```go
type Task struct {
    ID          string    `json:"id" db:"id"`
    JiraID      string    `json:"jira_id" db:"jira_id"`        // "OCPBUGS-1234" or "NO-JIRA"
    Title       string    `json:"title" db:"title"`
    Priority    Priority  `json:"priority" db:"priority"`
    Status      Status    `json:"status" db:"status"`
    Tags        []string  `json:"tags" db:"tags"`              // JSON array in DB
    Blockers    []string  `json:"blockers" db:"blockers"`      // JSON array in DB
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

### Link

Represents external resources related to a task (PRs, Slack threads, Jira tickets, etc.).

```go
type Link struct {
    ID       string   `json:"id" db:"id"`
    TaskID   string   `json:"task_id" db:"task_id"`
    Type     LinkType `json:"type" db:"type"`
    URL      string   `json:"url" db:"url"`
    Title    string   `json:"title" db:"title"`           // Title of PR, thread, etc.
    Status   string   `json:"status" db:"status"`         // Status of PR, ticket, etc.
    Metadata string   `json:"metadata" db:"metadata"`     // JSON for type-specific data
}
```

### Comment

User notes and comments attached to tasks.

```go
type Comment struct {
    ID        string    `json:"id" db:"id"`
    TaskID    string    `json:"task_id" db:"task_id"`
    Content   string    `json:"content" db:"content"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

## Enums and Constants

### Priority

```go
type Priority string

const (
    Minor    Priority = "minor"
    Normal   Priority = "normal" 
    High     Priority = "high"
    Critical Priority = "critical"
)
```

### Status

```go
type Status string

const (
    New        Status = "new"
    InProgress Status = "in_progress"
    Blocked    Status = "blocked"
    Done       Status = "done"
    Archived   Status = "archived"
)
```

**Status Flow:**
- `new` → `in_progress` → `done` → `archived`
- Any status can transition to `blocked` when blockers are added
- `blocked` → `in_progress` when blockers are resolved

### LinkType

```go
type LinkType string

const (
    PullRequest   LinkType = "pull_request"
    SlackThread   LinkType = "slack_thread"
    JiraTicket    LinkType = "jira_ticket"
    Documentation LinkType = "documentation"
    Other         LinkType = "other"
)
```

## Relationships

- **Task** ← 1:N → **Link**: A task can have multiple related links
- **Task** ← 1:N → **Comment**: A task can have multiple comments
- All relationships use foreign keys with CASCADE DELETE to maintain data integrity

## Business Rules

1. **JiraID**: Must be either a valid Jira ticket format (e.g., "OCPBUGS-1234") or "NO-JIRA"
2. **Status and Blockers**: When a task has non-empty blockers, it should typically be in "blocked" status
3. **Tags**: Stored as JSON array, case-insensitive for searching
4. **Timestamps**: All entities have creation timestamps, tasks also have update timestamps
5. **IDs**: Use UUIDs for all primary keys to ensure uniqueness