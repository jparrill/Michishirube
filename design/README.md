# Michishirube Design Documentation

<div align="center">
  <img src="../assets/michishirube-logo.png" alt="Michishirube Logo" width="200"/>
</div>

This directory contains the complete design documentation for Michishirube, a personal task organization tool for developers.

## Project Overview

Michishirube (道標 - "signpost" in Japanese) is a personal task management application designed to help developers organize their daily work including Jira tickets, pull requests, Slack conversations, and related resources.

## Design Documents

- [Data Model](./data-model.md) - Core data structures and relationships
- [Architecture](./architecture.md) - Project structure and component organization  
- [Storage](./storage.md) - Database design and schema
- [API](./api.md) - REST API endpoints and specifications
- [Frontend](./frontend.md) - User interface design and wireframes
- [Migrations](./migrations.md) - Automatic database migration system

## Technology Stack

- **Backend**: Go with native HTTP server
- **Frontend**: HTML templates + CSS/JS (no complex frameworks)
- **Database**: SQLite with automatic migrations
- **Deployment**: Docker + docker-compose

## Key Features

- Import tasks from Jira or create custom ones
- Link related resources (PRs, Slack threads, documentation)
- Track task status with custom workflow states
- Search and filter tasks efficiently
- Archive old tasks while maintaining searchability
- Personal comments and blocker tracking