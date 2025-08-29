# Error Logging System API Documentation

## Overview

The Error Logging System provides a comprehensive RESTful API for capturing, managing, and analyzing application errors in real-time. The system supports error aggregation, caching, and provides detailed analytics.

## Base URL

```
http://localhost:8080
```

## Authentication

All API endpoints (except health check) require API key authentication using the `X-API-Key` header.

```http
X-API-Key: your-api-key-here
```

**Default API Key for Development:**

```
test-api-key
```

## Response Format

All API responses follow a consistent JSON format:

### Success Response

```json
{
  "data": { ... },
  "status": "success"
}
```

### Error Response

```json
{
  "error": "Error description",
  "status": "error"
}
```

## Endpoints

### Health Check

#### GET /health

Check the API server status.

**Authentication:** Not required

**Response:**

```json
{
  "status": "ok",
  "timestamp": "2025-08-29T12:00:00Z"
}
```

---

### Error Management

#### POST /api/errors

Create a new error entry.

**Authentication:** Required

**Request Body:**

```json
{
  "level": "error",
  "message": "Database connection failed",
  "stack_trace": "Error: Connection timeout\n    at Database.connect(db.js:45)\n    at main(app.js:12)",
  "context": {
    "user_id": "12345",
    "session_id": "abc-def-ghi",
    "request_id": "req-789",
    "additional_data": "any value"
  },
  "source": "backend",
  "environment": "production",
  "url": "https://api.example.com/users"
}
```

**Parameters:**

- `level` (string, optional): Error level - `error`, `warning`, `info`, `debug`. Default: `error`
- `message` (string, required): Error message
- `stack_trace` (string, optional): Stack trace information
- `context` (object, optional): Additional context data
- `source` (string, required): Source of the error - `frontend`, `backend`, `api`, etc.
- `environment` (string, optional): Environment where error occurred. Default: `production`
- `url` (string, optional): URL where error occurred

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-08-29T12:00:00Z",
  "level": "error",
  "message": "Database connection failed",
  "stack_trace": "Error: Connection timeout...",
  "context": { ... },
  "source": "backend",
  "environment": "production",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.100",
  "url": "https://api.example.com/users",
  "fingerprint": "abc123def456",
  "resolved": false,
  "count": 1,
  "first_seen": "2025-08-29T12:00:00Z",
  "last_seen": "2025-08-29T12:00:00Z",
  "created_at": "2025-08-29T12:00:00Z",
  "updated_at": "2025-08-29T12:00:00Z"
}
```

---

#### GET /api/errors

Retrieve a list of errors with optional filtering and pagination.

**Authentication:** Required

**Query Parameters:**

- `limit` (integer, optional): Number of errors to return (1-100). Default: `50`
- `offset` (integer, optional): Number of errors to skip. Default: `0`
- `level` (string, optional): Filter by error level
- `source` (string, optional): Filter by error source

**Examples:**

```http
GET /api/errors?limit=20&offset=0
GET /api/errors?level=error&source=frontend
GET /api/errors?limit=10&level=warning
```

**Response:**

```json
{
  "errors": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "timestamp": "2025-08-29T12:00:00Z",
      "level": "error",
      "message": "Database connection failed",
      "source": "backend",
      "resolved": false,
      "count": 5
      // ... other fields
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 50
}
```

---

#### GET /api/errors/{id}

Retrieve a specific error by ID.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): Error ID

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-08-29T12:00:00Z",
  "level": "error",
  "message": "Database connection failed",
  "stack_trace": "Error: Connection timeout...",
  "context": {
    "user_id": "12345",
    "session_id": "abc-def-ghi"
  },
  "source": "backend",
  "environment": "production",
  "resolved": false,
  "count": 5
  // ... all other fields
}
```

**Error Responses:**

- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Error not found

---

#### PUT /api/errors/{id}/resolve

Mark an error as resolved.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): Error ID

**Response:**

```json
{
  "status": "resolved"
}
```

**Error Responses:**

- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Error not found

---

#### DELETE /api/errors/{id}

Delete a specific error.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): Error ID

**Response:**

- `204 No Content`: Error deleted successfully

**Error Responses:**

- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Error not found

---

### Analytics

#### GET /api/stats

Get error statistics and analytics.

**Authentication:** Required

**Response:**

```json
{
  "total_errors": 1250,
  "resolved_errors": 856,
  "errors_today": 23,
  "errors_this_week": 145,
  "errors_this_month": 567
}
```

---

## Error Levels

The system supports the following error levels:

- `error`: Critical errors that need immediate attention
- `warning`: Warning messages that might indicate potential issues
- `info`: Informational messages
- `debug`: Debug information for development

## Error Aggregation

The system automatically groups similar errors using a fingerprint algorithm based on:

- Error message
- Stack trace (if provided)

Similar errors are grouped together and the `count` field is incremented, with `first_seen` and `last_seen` timestamps updated accordingly.

## Caching

The API uses Redis for caching frequently accessed data:

- Error lists are cached for 2 minutes
- Statistics are cached for improved performance
- Cache is automatically invalidated when errors are created, updated, or deleted

## Rate Limiting

Currently, no rate limiting is implemented, but it's recommended for production use.

## Environment Variables

Configure the API using these environment variables:

```bash
# Database Configuration
DATABASE_URL=postgres://user:password@localhost:5432/error_logs?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379

# Server Configuration
PORT=8080
ENVIRONMENT=production
```

## Error Handling

The API returns appropriate HTTP status codes:

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `204 No Content`: Resource deleted successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Invalid or missing API key
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Usage Examples

### JavaScript/Frontend

```javascript
const API_KEY = "test-api-key";
const API_BASE = "http://localhost:8080/api";

// Create an error
async function logError(errorData) {
  const response = await fetch(`${API_BASE}/errors`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": API_KEY,
    },
    body: JSON.stringify(errorData),
  });

  return response.json();
}

// Get errors
async function getErrors(params = {}) {
  const query = new URLSearchParams(params);
  const response = await fetch(`${API_BASE}/errors?${query}`, {
    headers: {
      "X-API-Key": API_KEY,
    },
  });

  return response.json();
}

// Usage
logError({
  level: "error",
  message: "User authentication failed",
  source: "frontend",
  context: {
    user_id: "12345",
    page: "/login",
  },
});
```

### curl Examples

```bash
# Create an error
curl -X POST http://localhost:8080/api/errors \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "level": "error",
    "message": "Database connection failed",
    "source": "backend",
    "context": {"service": "user-service"}
  }'

# Get errors
curl -H "X-API-Key: test-api-key" \
  "http://localhost:8080/api/errors?limit=10&level=error"

# Get statistics
curl -H "X-API-Key: test-api-key" \
  "http://localhost:8080/api/stats"

# Resolve an error
curl -X PUT http://localhost:8080/api/errors/{error-id}/resolve \
  -H "X-API-Key: test-api-key"
```

### Go Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type ErrorRequest struct {
    Level   string                 `json:"level"`
    Message string                 `json:"message"`
    Source  string                 `json:"source"`
    Context map[string]interface{} `json:"context"`
}

func logError(errorReq ErrorRequest) error {
    data, _ := json.Marshal(errorReq)

    req, _ := http.NewRequest("POST", "http://localhost:8080/api/errors", bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "test-api-key")

    client := &http.Client{}
    resp, err := client.Do(req)
    defer resp.Body.Close()

    return err
}
```

## Docker Setup

The API can be run using Docker Compose. Ensure you have the following services:

```yaml
# docker-compose.yml
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://error_logs_user:error_logs_password@postgres:5432/error_logs?sslmode=disable
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=error_logs
      - POSTGRES_USER=error_logs_user
      - POSTGRES_PASSWORD=error_logs_password
    volumes:
      - ./database/schema.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
```

## Security Considerations

1. **API Keys**: Use strong, unique API keys in production
2. **HTTPS**: Always use HTTPS in production environments
3. **Rate Limiting**: Implement rate limiting to prevent abuse
4. **Input Validation**: The API validates all input data
5. **SQL Injection**: Uses parameterized queries to prevent SQL injection
6. **CORS**: Configured for cross-origin requests

## Performance Tips

1. Use pagination with reasonable limits (≤100 errors per request)
2. Cache statistics on the client side when possible
3. Use specific filters (level, source) to reduce response size
4. Monitor Redis cache hit rates for optimization

## Support

For issues or questions about the API, please check the application logs and ensure:

- Database connection is successful
- Redis is accessible
- Valid API key is provided
- Request format matches the documentation
