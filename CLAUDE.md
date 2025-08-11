# Claude Context - Michishirube Project

This file contains important context for Claude to understand the Michishirube project design and decisions made during development sessions.

## Project Overview

Michishirube (道標 - "signpost" in Japanese) is a personal task management application designed to replace Arc browser's fixed tabs system for organizing daily developer work.

### Purpose
- Organize daily developer tasks including Jira tickets, PRs, Slack conversations
- Provide a unified view of work items across different platforms
- Enable quick access to related resources and historical information
- Support personal workflow organization independent of existing tools

### User Profile
- Single user application (personal use only)
- Developer working with Jira ecosystem (OCPBUGS-*, etc.)
- Handles 5-20 tasks daily that may span multiple days
- Needs quick access to related links (PRs, Slack, documentation)

## Key Design Decisions

### Architecture Choices
- **Go with standard library**: No complex frameworks, using native `net/http`
- **SQLite database**: Perfect for single-user, portable, zero-configuration
- **Server-side rendering**: HTML templates + minimal JavaScript for simplicity
- **Docker deployment**: Easy to run anywhere, containerized for portability

### Data Model Insights
- **Task-centric design**: Everything revolves around Task entity
- **Flexible linking**: Support multiple link types (PR, Slack, Jira, docs)
- **Status workflow**: `new` → `in_progress` → `blocked` → `done` → `archived`
- **Priority system**: Matches Jira conventions (minor, normal, high, critical)
- **Blocking system**: Explicit blockers field tied to "blocked" status

### User Experience Priorities
- **Information density**: Show maximum relevant info without clutter
- **Quick actions**: Minimal clicks for common operations
- **Search-first**: Powerful search across all task attributes
- **Visual hierarchy**: Clear status and priority indicators
- **Keyboard shortcuts**: Developer-friendly quick navigation

## Important Implementation Notes

### Database Schema
- Use UUIDs for all primary keys
- JSON fields for arrays (tags, blockers) - stored as TEXT in SQLite
- Foreign key constraints with CASCADE DELETE for data integrity
- Comprehensive indexing for search performance

### API Design
- RESTful endpoints following OpenAPI 3.0 specification
- Consistent error responses with proper HTTP status codes
- Support for filtering, pagination, and search
- Separate endpoints for status updates (PATCH) vs full updates (PUT)

### Migration System
- **Fully automatic**: No user intervention required
- **Transparent**: Runs on application startup
- **Safe**: Atomic operations with rollback capability
- **Version tracking**: Prevents duplicate applications

### Frontend Approach
- **Progressive enhancement**: Works without JavaScript, enhanced with it
- **Responsive design**: Desktop-first but mobile-friendly
- **Accessibility**: Proper semantic HTML and keyboard navigation
- **Developer UX**: Keyboard shortcuts and quick filters

## Development Workflow

### Testing Strategy
- Always run lint and typecheck commands after code changes
- Look for existing test patterns in the codebase
- Check README or search codebase for testing approach

### Code Conventions
- Follow existing patterns in the codebase
- Check package.json (or go.mod) for available libraries
- Never assume library availability - always verify first
- Use existing imports and patterns from neighboring files

### Commit Standards
- Use conventional commit format
- Include comprehensive descriptions for major changes
- Always include Claude Code signature in commits

## Technology Stack

### Backend
- **Language**: Go (latest stable version)
- **HTTP Server**: Standard library `net/http`
- **Database**: SQLite with GORM or similar ORM
- **Migrations**: Custom migration system

### Frontend
- **Templates**: Go's `html/template`
- **Styling**: Custom CSS (no frameworks)
- **JavaScript**: Vanilla JS for interactions
- **Icons**: Unicode/emoji or lightweight icon font

### DevOps
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: docker-compose for development
- **Build**: Makefile for common tasks

## File Structure Reference

```
michishirube/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── models/                 # Business entities
│   ├── storage/                # Data persistence layer
│   ├── handlers/               # HTTP request handlers
│   └── server/                 # HTTP server setup
├── web/
│   ├── static/                 # CSS, JS, assets
│   └── templates/              # HTML templates
├── design/                     # Complete design documentation
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Common Commands

When working on this project:

```bash
# Development
make build          # Build the application
make run           # Run in development mode
make test          # Run tests
make lint          # Run linting

# Docker
docker-compose up  # Run with docker-compose
make docker-build  # Build docker image

# Database
# Migrations run automatically on startup
# No manual database commands needed
```

## Future Considerations

### Potential Enhancements
- **Jira integration**: Automatic fetching of ticket status/details
- **GitHub integration**: PR status updates
- **Export functionality**: Backup and data portability
- **Themes**: Dark mode and customization options
- **Authentication**: If multi-user support needed

### Scalability Notes
- Current design supports hundreds of tasks efficiently
- SQLite performance adequate for single-user workloads
- Migration to PostgreSQL straightforward if needed
- API design allows for frontend framework adoption later

## Session Context

This project was designed from scratch in a single session, starting with user requirements and progressing through complete architectural design. All design decisions were made collaboratively with focus on simplicity, usability, and maintainability.

The user specifically wanted:
- Replacement for Arc browser's tab organization system
- Web-based interface accessible from any browser
- Support for both Jira tickets and custom tasks
- Rich linking to external resources
- Search and archival capabilities
- Personal workflow optimization

All code and documentation should be in English, though development discussions may be in Spanish.