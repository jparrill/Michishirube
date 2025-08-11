# Architecture

This document describes the overall architecture and project structure of Michishirube.

## Project Structure

```
michishirube/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management (DB, port, etc.)
│   ├── models/
│   │   ├── task.go              # Task struct and methods
│   │   ├── link.go              # Link struct and methods  
│   │   └── comment.go           # Comment struct and methods
│   ├── storage/
│   │   ├── interface.go         # Storage interface definition
│   │   └── sqlite/
│   │       ├── sqlite.go        # SQLite implementation
│   │       └── migrations.go    # Database migrations
│   ├── handlers/
│   │   ├── tasks.go             # Task CRUD operations
│   │   ├── links.go             # Link CRUD operations
│   │   ├── comments.go          # Comment CRUD operations
│   │   └── search.go            # Search functionality
│   └── server/
│       └── server.go            # HTTP server setup and routing
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── styles.css       # Application styles
│   │   ├── js/
│   │   │   └── app.js           # Frontend JavaScript
│   │   └── assets/
│   │       └── logo.png         # Static assets
│   └── templates/
│       ├── layout.html          # Base template
│       ├── index.html           # Main dashboard
│       ├── task.html            # Task detail view
│       └── new-task.html        # New task form
├── design/                      # Design documentation
├── Dockerfile                   # Docker container definition
├── docker-compose.yml          # Multi-container setup
├── Makefile                     # Build and development commands
├── go.mod                       # Go module definition
└── go.sum                       # Go module checksums
```

## Architectural Layers

### 1. Presentation Layer (`cmd/server`, `web/`)
- **Entry Point**: `cmd/server/main.go` - Application bootstrap
- **Web Server**: HTTP server setup and middleware
- **Templates**: HTML templates for the web interface
- **Static Assets**: CSS, JavaScript, and images

### 2. Handler Layer (`internal/handlers/`)
- **HTTP Handlers**: REST API endpoints
- **Request/Response**: JSON serialization and validation
- **Routing**: URL pattern matching and method handling

### 3. Business Logic Layer (`internal/models/`)
- **Domain Models**: Core business entities (Task, Link, Comment)
- **Business Rules**: Validation and business logic
- **Domain Services**: Cross-entity operations

### 4. Storage Layer (`internal/storage/`)
- **Interface**: Abstract storage operations
- **Implementation**: SQLite-specific data access
- **Migrations**: Database schema evolution

### 5. Configuration (`internal/config/`)
- **Environment**: Configuration from env vars and files
- **Database**: Connection settings
- **Server**: Port and middleware configuration

## Design Principles

### Dependency Inversion
- Higher layers depend on abstractions, not implementations
- Storage interface allows for easy database switching
- Handlers depend on storage interface, not concrete implementation

### Clean Architecture
- Domain models are independent of external concerns
- Business logic is separated from delivery mechanisms
- External dependencies (database, web) are in outer layers

### Single Responsibility
- Each package has a clear, focused purpose
- Handlers only handle HTTP concerns
- Storage only handles data persistence
- Models only contain business logic

## Package Dependencies

```
cmd/server → internal/server → internal/handlers → internal/models
                            ↘ internal/storage ↗
                              internal/config
```

## Key Design Decisions

1. **Standard Library Focus**: Use Go's standard `net/http` instead of frameworks for simplicity
2. **Interface-Based Storage**: Storage operations defined by interface for testability
3. **Template-Based Frontend**: Server-side rendering with minimal JavaScript
4. **Internal Package**: All business logic in `internal/` to prevent external imports
5. **Embedded Assets**: Static files can be embedded in binary for distribution

## Extensibility Points

- **Storage Backends**: New databases can be added by implementing the storage interface
- **Authentication**: Can be added as middleware in the server layer
- **API Versions**: Can be handled through routing patterns
- **Export Formats**: Can be added as new handlers with different content types