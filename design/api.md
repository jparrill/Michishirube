# API Design

This document describes the REST API endpoints for Michishirube.

## OpenAPI Specification

The complete API specification is available in OpenAPI 3.0 format: [openapi.yaml](./openapi.yaml)

You can view the interactive documentation by:
1. Using Swagger UI with the openapi.yaml file
2. Importing the spec into Postman or similar API clients
3. Using online viewers like editor.swagger.io

## Base URL

```
http://localhost:8080/api
```

## Authentication

For initial version, no authentication is required as this is a personal tool. Future versions may add basic auth or token-based authentication.

## Content Type

All API endpoints expect and return `application/json` unless otherwise specified.

## Error Responses

Standard HTTP status codes are used. Error responses have the following format:

```json
{
    "error": "Description of the error",
    "code": "ERROR_CODE"
}
```

## Endpoints

### Tasks

#### List Tasks
```
GET /api/tasks
```

**Query Parameters:**
- `status` (string, optional): Filter by status (comma-separated for multiple)
- `priority` (string, optional): Filter by priority (comma-separated for multiple)
- `tags` (string, optional): Filter by tags (comma-separated)
- `include_archived` (boolean, optional): Include archived tasks (default: false)
- `limit` (int, optional): Maximum number of results (default: 50)
- `offset` (int, optional): Number of results to skip (default: 0)

**Example:**
```
GET /api/tasks?status=new,in_progress&priority=high,critical&limit=20
```

**Response:**
```json
{
    "tasks": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "jira_id": "OCPBUGS-1234",
            "title": "Fix memory leak in pod controller",
            "priority": "high",
            "status": "blocked",
            "tags": ["k8s", "memory"],
            "blockers": ["Waiting for review from @team-lead"],
            "created_at": "2024-01-15T10:30:00Z",
            "updated_at": "2024-01-15T14:20:00Z"
        }
    ],
    "total": 42,
    "limit": 20,
    "offset": 0
}
```

#### Create Task
```
POST /api/tasks
```

**Request Body:**
```json
{
    "jira_id": "OCPBUGS-5678",
    "title": "Implement new feature",
    "priority": "normal",
    "tags": ["feature", "api"],
    "blockers": []
}
```

**Response:**
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "jira_id": "OCPBUGS-5678",
    "title": "Implement new feature",
    "priority": "normal",
    "status": "new",
    "tags": ["feature", "api"],
    "blockers": [],
    "created_at": "2024-01-15T15:30:00Z",
    "updated_at": "2024-01-15T15:30:00Z"
}
```

#### Get Task
```
GET /api/tasks/{id}
```

**Response:**
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "jira_id": "OCPBUGS-1234",
    "title": "Fix memory leak in pod controller",
    "priority": "high",
    "status": "blocked",
    "tags": ["k8s", "memory"],
    "blockers": ["Waiting for review from @team-lead"],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T14:20:00Z",
    "links": [
        {
            "id": "link-id-1",
            "type": "pull_request",
            "url": "https://github.com/org/repo/pull/456",
            "title": "Fix memory leak",
            "status": "draft"
        }
    ],
    "comments": [
        {
            "id": "comment-id-1",
            "content": "Need to investigate heap allocation patterns",
            "created_at": "2024-01-15T11:00:00Z"
        }
    ]
}
```

#### Update Task
```
PUT /api/tasks/{id}
```

**Request Body:** (Full task object)
```json
{
    "jira_id": "OCPBUGS-1234",
    "title": "Fix memory leak in pod controller - Updated",
    "priority": "critical",
    "status": "in_progress",
    "tags": ["k8s", "memory", "urgent"],
    "blockers": []
}
```

#### Update Task Status
```
PATCH /api/tasks/{id}/status
```

**Request Body:**
```json
{
    "status": "done"
}
```

#### Delete Task
```
DELETE /api/tasks/{id}
```

**Response:** `204 No Content`

### Links

#### Get Links for Task
```
GET /api/tasks/{taskId}/links
```

**Response:**
```json
{
    "links": [
        {
            "id": "link-id-1",
            "task_id": "task-id-1",
            "type": "pull_request",
            "url": "https://github.com/org/repo/pull/456",
            "title": "Fix memory leak",
            "status": "merged",
            "metadata": "{\"pr_number\": 456, \"author\": \"user123\"}"
        }
    ]
}
```

#### Add Link to Task
```
POST /api/tasks/{taskId}/links
```

**Request Body:**
```json
{
    "type": "slack_thread",
    "url": "https://company.slack.com/archives/C1234/p1234567890",
    "title": "Discussion about memory issue",
    "status": "active"
}
```

#### Update Link
```
PUT /api/links/{id}
```

#### Delete Link
```
DELETE /api/links/{id}
```

### Comments

#### Get Comments for Task
```
GET /api/tasks/{taskId}/comments
```

**Response:**
```json
{
    "comments": [
        {
            "id": "comment-id-1",
            "task_id": "task-id-1",
            "content": "Need to investigate heap allocation patterns",
            "created_at": "2024-01-15T11:00:00Z"
        }
    ]
}
```

#### Add Comment to Task
```
POST /api/tasks/{taskId}/comments
```

**Request Body:**
```json
{
    "content": "Found the root cause in the controller reconcile loop"
}
```

#### Delete Comment
```
DELETE /api/comments/{id}
```

### Search

#### Search Tasks
```
GET /api/search
```

**Query Parameters:**
- `q` (string, required): Search query
- `include_archived` (boolean, optional): Include archived tasks (default: false)
- `limit` (int, optional): Maximum number of results (default: 20)

**Example:**
```
GET /api/search?q=OCPBUGS-1234&include_archived=true
```

**Response:**
```json
{
    "tasks": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "jira_id": "OCPBUGS-1234",
            "title": "Fix memory leak in pod controller",
            "priority": "high",
            "status": "blocked",
            "tags": ["k8s", "memory"],
            "blockers": ["Waiting for review from @team-lead"],
            "created_at": "2024-01-15T10:30:00Z",
            "updated_at": "2024-01-15T14:20:00Z"
        }
    ],
    "query": "OCPBUGS-1234",
    "total": 1
}
```

#### Get Search Suggestions
```
GET /api/search/suggestions
```

**Query Parameters:**
- `type` (string, optional): Type of suggestions (tags, jira_ids)

**Response:**
```json
{
    "suggestions": {
        "tags": ["k8s", "memory", "frontend", "bug"],
        "jira_ids": ["OCPBUGS-1234", "OCPBUGS-5678"]
    }
}
```

## Web UI Routes

These routes serve HTML pages for the web interface:

```
GET /                    # Main dashboard
GET /task/{id}          # Task detail view
GET /new                # New task form
```

## Status Codes

- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `204 No Content` - Successful request with no response body
- `400 Bad Request` - Invalid request format or parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error