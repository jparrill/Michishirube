# Michishirube (道標)

<img src="assets/michishirube-logo.png" align="left" width="120" style="margin-right: 20px;">

> Personal task management application for developers

Michishirube (Japanese for "signpost") is a personal task management tool designed specifically for developers. It helps organize daily work including Jira tickets, pull requests, Slack conversations, and provides a unified view of work items across different platforms.

[![CI/CD Pipeline](https://github.com/jparrill/Michishirube/actions/workflows/ci.yml/badge.svg)](https://github.com/jparrill/Michishirube/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/jparrill/Michishirube/branch/main/graph/badge.svg)](https://codecov.io/gh/jparrill/Michishirube)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jparrill/Michishirube)](https://github.com/jparrill/Michishirube/blob/main/go.mod)
[![Go Documentation](https://godoc.org/github.com/jparrill/Michishirube?status.svg)](https://godoc.org/github.com/jparrill/Michishirube)
[![Release](https://img.shields.io/github/v/release/jparrill/Michishirube)](https://github.com/jparrill/Michishirube/releases)
[![License](https://img.shields.io/github/license/jparrill/Michishirube)](https://github.com/jparrill/Michishirube/blob/main/LICENSE)
[![Docker Image (Quay.io)](https://img.shields.io/badge/quay.io-jparrill%2Fmichishirube-blue)](https://quay.io/repository/jparrill/michishirube)
[![Docker Image (Docker Hub)](https://img.shields.io/badge/docker.io-padajuan%2Fmichishirube-blue)](https://hub.docker.com/r/padajuan/michishirube)

## Features

- **Task Management**: Create, update, and track tasks with priorities and statuses
- **Jira Integration**: Support for JIRA ticket IDs and workflows
- **Link Management**: Associate pull requests, Slack threads, documentation, and other resources with tasks
- **Comments**: Add notes and updates to track progress
- **Search**: Powerful search across all task attributes
- **Status Reports**: Automatic generation of "working on", "next up", and "blockers" reports
- **API Documentation**: Complete REST API with Swagger UI
- **Single User**: Designed for personal productivity (not multi-user)

## Quick Start

### Download

Download the latest release for your platform from the [releases page](https://github.com/jparrill/michishirube/releases).

### Installation Options

#### Option 1: Binary Installation
1. Extract the downloaded archive
2. Run the binary:
   ```bash
   ./michishirube
   ```
3. Open your browser to http://localhost:8080

#### Option 2: Docker (Recommended)
```bash
# Run with docker-compose (production)
docker-compose up -d

# Or run directly with Docker
docker run -d -p 8080:8080 -v michishirube_data:/data quay.io/jparrill/michishirube:latest
```

#### Option 3: Building from Source
```bash
# Clone the repository
git clone https://github.com/jparrill/michishirube.git
cd michishirube

# Install dependencies
go mod download

# Build the application
make build

# Run the application
./build/michishirube
```

### Creating Releases

```bash
# Create a version tag
git tag v1.0.0

# Check configuration
make release-check

# Test with snapshot build
make release-snapshot

# Create and publish release
make release
```

The release process uses GoReleaser to create multiarch binaries for:
- **macOS**: Intel (amd64) and Apple Silicon (arm64) with CGO SQLite driver
- **Linux**: Intel (amd64) and ARM64 (arm64) with pure Go SQLite driver

### CI/CD Pipeline

The project uses GitHub Actions for automated testing and releases:

- **Continuous Integration**: Runs on every push and pull request
  - Tests across all packages with coverage reporting
  - Security scanning with Trivy and Gosec
  - Linting with golangci-lint
  - Binary build and smoke testing
  - GoReleaser configuration validation

- **Snapshot Builds**: Created on every push to main branch
  - Multiarch binaries available as artifacts
  - Useful for testing latest changes

- **Release Automation**: Triggered by version tags (e.g., `v1.0.0`)
  - Automatic GitHub releases with changelog
  - Multiarch binary distribution
  - Checksums and signatures for security

- **Dependency Management**: Automated updates via Dependabot
  - Weekly Go module updates
  - GitHub Actions version updates

### Docker Deployment

The project includes comprehensive Docker support with multiarch builds:

```bash
# Quick start with docker-compose
cp .env.example .env  # Configure environment
docker-compose up -d  # Start production services

# Development mode with hot reload
docker-compose --profile dev up --build

# Multiarch build for distribution
make docker-multiarch
```

**Environment Configuration:**
Create a `.env` file from `.env.example` to customize:
- `PORT`: Application port (default: 8080)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `DB_PATH`: Database file path (persisted in Docker volume)

**Docker Commands:**
- `make docker-build`: Single architecture build
- `make docker-multiarch`: Multi-platform build (amd64, arm64)
- `make docker-up`: Start production environment
- `make docker-dev`: Start development environment with hot reload
- `make docker-down`: Stop all services
- `make docker-clean`: Clean all Docker resources

## Usage

### Web Interface

- **Dashboard**: View all tasks with filtering and search
- **Create Task**: Add new tasks with JIRA IDs, priorities, and tags
- **Task Details**: View task with associated links and comments
- **Search**: Find tasks by title, description, tags, or JIRA ID

### API Interface

The application provides a complete REST API documented with Swagger:

- **API Documentation**: http://localhost:8080/docs
- **OpenAPI Specification**: http://localhost:8080/openapi.yaml

#### Key Endpoints

- `GET /api/tasks` - List and filter tasks
- `POST /api/tasks` - Create new task
- `GET /api/tasks/{id}` - Get task details
- `PATCH /api/tasks/{id}` - Update task fields
- `POST /api/links` - Add links to tasks
- `POST /api/comments` - Add comments to tasks
- `GET /api/report` - Generate status report

## Configuration

Michishirube can be configured via environment variables:

- `PORT`: Server port (default: 8080)
- `DB_PATH`: SQLite database path (default: ./michishirube.db)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## Development

### Prerequisites

- Go 1.24.6 or later
- SQLite3 (for CGO)

### Available Commands

```bash
# Development
make build          # Build the application
make run            # Run in development mode
make test           # Run complete test suite
make test-unit      # Run only unit tests
make docs           # Generate API documentation

# Release Management
make release-check     # Check GoReleaser configuration
make release-snapshot  # Create snapshot build for development
make release          # Create and publish a release (requires git tag)

# Help
make test-help     # Show all test-related commands
make release-help  # Show all release-related commands
```

See `make test-help` and `make release-help` for complete command lists.

### Project Structure

```
michishirube/
├── cmd/server/         # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP request handlers
│   ├── models/         # Business entities and DTOs
│   ├── server/         # HTTP server setup
│   └── storage/        # Data persistence layer
├── web/
│   ├── static/         # CSS, JS, assets
│   └── templates/      # HTML templates
└── docs/              # Generated API documentation
```

## Technology Stack

- **Backend**: Go with standard library HTTP server
- **Database**: SQLite with automatic migrations
  - CGO driver (`github.com/mattn/go-sqlite3`) for macOS builds
  - Pure Go driver (`modernc.org/sqlite`) for Linux cross-compilation
- **Frontend**: Server-side rendered HTML with vanilla JavaScript
- **API Documentation**: Swagger/OpenAPI 3.0 with Swaggo
- **Testing**: Comprehensive test suite with fixtures
- **Release Management**: GoReleaser for multiarch binary distribution
- **CI/CD**: GitHub Actions with automated testing, security scanning, and releases

## Contributing

This is a personal project, but feedback and suggestions are welcome via GitHub issues.

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built with Go's excellent standard library
- Uses SQLite for zero-configuration persistence
- Swagger UI for API documentation
- Inspired by the need for better personal task organization

---

*Michishirube - Your personal signpost for navigating daily development work.*