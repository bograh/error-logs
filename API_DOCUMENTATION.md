# Error Logging System API Documentation

## Overview

The Error Logging System provides a comprehensive RESTful API for capturing, managing, and analyzing application errors in real-time. The system supports error aggregation, caching, advanced analytics, monitoring, alerting, and team management capabilities.

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
  "data": {
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
  },
  "status": "success"
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
  "data": {
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
  },
  "status": "success"
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
  "data": {
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
  },
  "status": "success"
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
  "data": {
    "status": "resolved"
  },
  "status": "success"
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
  "data": {
    "total_errors": 1250,
    "resolved_errors": 856,
    "errors_today": 23,
    "errors_this_week": 145,
    "errors_this_month": 567,
    "error_rate_24h": 2.3,
    "resolution_rate": 68.5,
    "avg_resolution_time": "2h 15m"
  },
  "status": "success"
}
```

---

### Analytics

#### GET /api/analytics/trends

Get error trend data for charts and analytics.

**Authentication:** Required

**Query Parameters:**

- `period` (string, optional): Time period - `day`, `week`, `month`, `year`. Default: `week`
- `group_by` (string, optional): Group data by - `hour`, `day`, `week`, `month`. Default: `day`

**Examples:**

```http
GET /api/analytics/trends?period=week&group_by=day
GET /api/analytics/trends?period=month&group_by=week
```

**Response:**

```json
{
  "data": {
    "period": "week",
    "data_points": [
      {
        "timestamp": "2025-08-22T00:00:00Z",
        "error_count": 45,
        "resolved_count": 32,
        "critical_count": 3
      },
      {
        "timestamp": "2025-08-23T00:00:00Z",
        "error_count": 52,
        "resolved_count": 38,
        "critical_count": 1
      }
    ]
  },
  "status": "success"
}
```

---

#### GET /api/analytics/performance

Get system performance metrics.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "avg_response_time": 245,
    "error_rate_percent": 0.8,
    "throughput_rpm": 1200,
    "availability_percent": 99.95,
    "performance_score": 8.7
  },
  "status": "success"
}
```

---

### Monitoring

#### GET /api/monitoring/services

Get health status of all monitored services.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "services": [
      {
        "name": "Database",
        "status": "healthy",
        "uptime_percent": 99.8,
        "response_time_ms": 12,
        "last_checked": "2025-08-29T12:00:00Z",
        "details": {
          "connections": 45,
          "max_connections": 100
        }
      },
      {
        "name": "Cache Service",
        "status": "healthy",
        "uptime_percent": 99.95,
        "response_time_ms": 3,
        "last_checked": "2025-08-29T12:00:00Z"
      }
    ],
    "overall_health": "healthy"
  },
  "status": "success"
}
```

---

#### GET /api/monitoring/metrics

Get system performance metrics.

**Authentication:** Required

**Query Parameters:**

- `timeframe` (string, optional): Time frame - `1h`, `6h`, `24h`, `7d`, `30d`. Default: `1h`

**Response:**

```json
{
  "data": {
    "cpu_usage_percent": 65.2,
    "memory_usage_percent": 78.1,
    "disk_usage_percent": 45.7,
    "network_io": {
      "bytes_in": 1024000,
      "bytes_out": 2048000
    },
    "active_connections": 45,
    "requests_per_minute": 1200
  },
  "status": "success"
}
```

---

#### GET /api/monitoring/uptime

Get historical uptime data.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "current_uptime_hours": 720.5,
    "uptime_percent_24h": 100.0,
    "uptime_percent_7d": 99.8,
    "uptime_percent_30d": 99.95,
    "incidents_count": 2,
    "last_downtime": "2025-08-10T14:30:00Z"
  },
  "status": "success"
}
```

---

### Alert Management

#### GET /api/alerts/rules

Get all alert rules.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "rules": [
      {
        "id": "rule-1",
        "name": "High Error Rate",
        "condition": "error_count > 50 in 5 minutes",
        "threshold": 50,
        "time_window": "5m",
        "enabled": true,
        "notifications": ["email", "slack"],
        "last_triggered": "2025-08-15T10:30:00Z",
        "created_at": "2025-08-10T09:00:00Z",
        "updated_at": "2025-08-15T10:30:00Z"
      }
    ]
  },
  "status": "success"
}
```

---

#### POST /api/alerts/rules

Create a new alert rule.

**Authentication:** Required

**Request Body:**

```json
{
  "name": "High Error Rate",
  "condition": "error_count > threshold in time_window",
  "threshold": 50,
  "time_window": "5m",
  "notifications": ["email", "slack"],
  "enabled": true
}
```

**Response:**

```json
{
  "data": {
    "id": "rule-2",
    "name": "High Error Rate",
    "condition": "error_count > threshold in time_window",
    "threshold": 50,
    "time_window": "5m",
    "enabled": true,
    "notifications": ["email", "slack"],
    "last_triggered": null,
    "created_at": "2025-08-29T12:00:00Z",
    "updated_at": "2025-08-29T12:00:00Z"
  },
  "status": "success"
}
```

---

#### PUT /api/alerts/rules/{id}

Update an alert rule.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): Alert rule ID

**Request Body:** Same as POST /api/alerts/rules

**Response:** Updated alert rule object

---

#### DELETE /api/alerts/rules/{id}

Delete an alert rule.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): Alert rule ID

**Response:**

- `204 No Content`: Alert rule deleted successfully

---

#### GET /api/alerts/incidents

Get all incidents.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "incidents": [
      {
        "id": "incident-1",
        "title": "Database Connection Timeout",
        "severity": "critical",
        "status": "investigating",
        "description": "Multiple database connection timeouts detected",
        "assigned_to": "user-123",
        "created_at": "2025-08-29T14:30:00Z",
        "updated_at": "2025-08-29T14:45:00Z"
      }
    ]
  },
  "status": "success"
}
```

---

#### POST /api/alerts/incidents

Create a new incident.

**Authentication:** Required

**Request Body:**

```json
{
  "title": "Database Connection Timeout",
  "severity": "critical",
  "description": "Multiple database connection timeouts detected",
  "assigned_to": "user-123"
}
```

**Response:** Created incident object

---

#### PUT /api/alerts/incidents/{id}

Update an incident.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): Incident ID

**Request Body:** Same as POST /api/alerts/incidents

**Response:** Updated incident object

---

### Settings & Configuration

#### GET /api/settings/api-keys

Get list of API keys.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "api_keys": [
      {
        "id": "key-1",
        "name": "Production API Key",
        "key_preview": "sk_****7890",
        "permissions": ["read", "write"],
        "expires_at": "2025-12-31T23:59:59Z",
        "last_used": "2025-08-29T10:30:00Z",
        "created_at": "2025-08-10T09:00:00Z"
      }
    ]
  },
  "status": "success"
}
```

---

#### POST /api/settings/api-keys

Create a new API key.

**Authentication:** Required

**Request Body:**

```json
{
  "name": "Development Key",
  "permissions": ["read"],
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:**

```json
{
  "data": {
    "id": "key-2",
    "name": "Development Key",
    "api_key": "sk_live_abc123def456...",
    "permissions": ["read"],
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2025-08-29T12:00:00Z"
  },
  "status": "success"
}
```

**Note:** The actual API key is only shown once during creation.

---

#### DELETE /api/settings/api-keys/{id}

Revoke an API key.

**Authentication:** Required

**Parameters:**

- `id` (UUID, required): API key ID

**Response:**

- `204 No Content`: API key revoked successfully

---

#### GET /api/settings/team

Get team members.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "members": [
      {
        "id": "user-1",
        "name": "John Doe",
        "email": "john@company.com",
        "role": "developer",
        "status": "active",
        "last_active": "2025-08-29T10:30:00Z",
        "created_at": "2025-08-10T09:00:00Z"
      }
    ]
  },
  "status": "success"
}
```

---

#### POST /api/settings/team/invite

Invite a team member.

**Authentication:** Required

**Request Body:**

```json
{
  "email": "newuser@company.com",
  "role": "viewer"
}
```

**Response:** Created team member object

---

#### GET /api/settings/integrations

Get integration status.

**Authentication:** Required

**Response:**

```json
{
  "data": {
    "integrations": [
      {
        "name": "slack",
        "status": "connected",
        "config": {
          "webhook_url": "https://hooks.slack.com/...",
          "channel": "#alerts"
        },
        "last_test": "2025-08-29T10:30:00Z"
      },
      {
        "name": "email",
        "status": "configured",
        "config": {
          "smtp_server": "smtp.example.com",
          "from_email": "alerts@example.com"
        },
        "last_test": null
      }
    ]
  },
  "status": "success"
}
```

---

## Data Models

### Error Object

```typescript
interface BackendError {
  id: string;
  timestamp: string;
  level: "error" | "warning" | "info" | "debug";
  message: string;
  stack_trace?: string;
  context?: Record<string, any>;
  source: string;
  environment?: string;
  user_agent?: string;
  ip_address?: string;
  url?: string;
  fingerprint?: string;
  resolved: boolean;
  count: number;
  first_seen: string;
  last_seen: string;
  created_at: string;
  updated_at: string;
}
```

### Service Health Object

```typescript
interface ServiceHealth {
  name: string;
  status: "healthy" | "unhealthy" | "unknown";
  uptime_percent: number;
  response_time_ms: number;
  last_checked: string;
  details?: Record<string, any>;
}
```

### Alert Rule Object

```typescript
interface AlertRule {
  id: string;
  name: string;
  condition: string;
  threshold: number;
  time_window: string;
  enabled: boolean;
  notifications: string[];
  last_triggered?: string;
  created_at: string;
  updated_at: string;
}
```

### API Key Object

```typescript
interface APIKey {
  id: string;
  name: string;
  key_preview: string;
  permissions: string[];
  expires_at?: string;
  last_used?: string;
  created_at: string;
}
```

---

## Error Levels

The system supports the following error levels:

- `error`: Critical errors that need immediate attention
- `warning`: Warning messages that might indicate potential issues
- `info`: Informational messages
- `debug`: Debug information for development

## User Roles

The system supports role-based access control:

- `owner`: Full access to all features and settings
- `admin`: Manage settings, view all data, manage team
- `developer`: View errors, manage alert rules, access analytics
- `viewer`: Read-only access to errors and analytics

## Alert Severities

Incidents and alerts can have different severity levels:

- `low`: Minor issues that don't require immediate attention
- `medium`: Standard issues that should be addressed
- `high`: Important issues requiring prompt attention
- `critical`: Critical issues requiring immediate response

## Notification Types

Supported notification channels for alerts:

- `email`: Email notifications
- `slack`: Slack webhook notifications
- `webhook`: Custom webhook notifications
- `sms`: SMS notifications (if configured)

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

Currently, no rate limiting is implemented, but it's recommended for production use. The API is designed to handle the following expected limits:

- Error logging: 1000 requests/minute per API key
- Analytics queries: 100 requests/minute per API key
- General API: 500 requests/minute per API key

## Caching Strategy

The API uses Redis for caching to improve performance:

- Error lists: Cached for 2 minutes
- Statistics: Cached for 5 minutes
- Analytics trends: Cached for 5 minutes
- Service health: Cached for 30 seconds
- System metrics: Cached for 30 seconds
- Performance metrics: Cached for 1 minute
- Uptime data: Cached for 5 minutes

Cache is automatically invalidated when data changes (errors created, resolved, or deleted).

## Security Features

1. **API Key Authentication**: All endpoints require valid API keys
2. **Role-based Access Control**: Different permission levels for team members
3. **Input Validation**: All input data is validated before processing
4. **SQL Injection Protection**: Uses parameterized queries
5. **Data Sanitization**: Sensitive data is automatically sanitized
6. **CORS Configuration**: Properly configured for cross-origin requests
7. **Rate Limiting Ready**: Infrastructure prepared for rate limiting implementation

## Database Schema

The system uses PostgreSQL with the following main tables:

- `errors`: Main error storage with fingerprinting and aggregation
- `api_keys`: API key management with permissions
- `alert_rules`: Alert rule definitions and configuration
- `incidents`: Incident tracking and management
- `team_members`: Team member management with roles
- `projects`: Multi-project support (future feature)

Required indexes are automatically created for optimal performance.

## Advanced Features

### Error Fingerprinting & Aggregation

The system automatically groups similar errors using a fingerprint algorithm based on:

- Normalized error message
- Stack trace patterns
- Source location

Similar errors are aggregated with count tracking and first/last seen timestamps.

### Real-time Capabilities

- Background queue processing for high-volume error ingestion
- Redis-based caching for fast response times
- Prepared for WebSocket support for real-time updates

### Monitoring & Alerting

- Comprehensive service health monitoring
- Customizable alert rules with multiple notification channels
- Incident management with severity tracking
- Performance metrics collection

### Team Collaboration

- Role-based access control
- Team member invitation system
- API key management with granular permissions
- Integration support for external services

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

// Get analytics trends
async function getTrends(period = "week", groupBy = "day") {
  const response = await fetch(
    `${API_BASE}/analytics/trends?period=${period}&group_by=${groupBy}`,
    {
      headers: {
        "X-API-Key": API_KEY,
      },
    }
  );

  return response.json();
}

// Get service health
async function getServiceHealth() {
  const response = await fetch(`${API_BASE}/monitoring/services`, {
    headers: {
      "X-API-Key": API_KEY,
    },
  });

  return response.json();
}

// Create alert rule
async function createAlertRule(ruleData) {
  const response = await fetch(`${API_BASE}/alerts/rules`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": API_KEY,
    },
    body: JSON.stringify(ruleData),
  });

  return response.json();
}

// Usage examples
logError({
  level: "error",
  message: "User authentication failed",
  source: "frontend",
  context: {
    user_id: "12345",
    page: "/login",
  },
});

// Get performance metrics
getTrends("month", "week").then((data) => {
  console.log("Trend data:", data);
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

# Get analytics trends
curl -H "X-API-Key: test-api-key" \
  "http://localhost:8080/api/analytics/trends?period=week&group_by=day"

# Get service health
curl -H "X-API-Key: test-api-key" \
  "http://localhost:8080/api/monitoring/services"

# Get system metrics
curl -H "X-API-Key: test-api-key" \
  "http://localhost:8080/api/monitoring/metrics?timeframe=24h"

# Create alert rule
curl -X POST http://localhost:8080/api/alerts/rules \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "name": "High Error Rate",
    "condition": "error_count > threshold in time_window",
    "threshold": 50,
    "time_window": "5m",
    "notifications": ["email"],
    "enabled": true
  }'

# Create API key
curl -X POST http://localhost:8080/api/settings/api-keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "name": "Development Key",
    "permissions": ["read"]
  }'

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

### Troubleshooting

**Common Issues:**

1. **401 Unauthorized**

   - Check API key is included in `X-API-Key` header
   - Verify API key is valid and active
   - Ensure API key has required permissions

2. **500 Internal Server Error**

   - Check database connectivity
   - Verify Redis is running and accessible
   - Review application logs for specific error details

3. **Slow Response Times**

   - Monitor Redis cache hit rates
   - Check database query performance
   - Consider increasing cache TTL for frequently accessed data

4. **High Memory Usage**
   - Review Redis memory usage and eviction policies
   - Monitor error ingestion rates
   - Consider implementing data retention policies

### API Endpoint Summary

| Endpoint                     | Method              | Purpose             | Auth Required |
| ---------------------------- | ------------------- | ------------------- | ------------- |
| `/health`                    | GET                 | Health check        | No            |
| `/api/errors`                | GET                 | List errors         | Yes           |
| `/api/errors`                | POST                | Create error        | Yes           |
| `/api/errors/{id}`           | GET                 | Get error           | Yes           |
| `/api/errors/{id}/resolve`   | PUT                 | Resolve error       | Yes           |
| `/api/errors/{id}`           | DELETE              | Delete error        | Yes           |
| `/api/stats`                 | GET                 | Get statistics      | Yes           |
| `/api/analytics/trends`      | GET                 | Get trends          | Yes           |
| `/api/analytics/performance` | GET                 | Performance metrics | Yes           |
| `/api/monitoring/services`   | GET                 | Service health      | Yes           |
| `/api/monitoring/metrics`    | GET                 | System metrics      | Yes           |
| `/api/monitoring/uptime`     | GET                 | Uptime data         | Yes           |
| `/api/alerts/rules`          | GET/POST/PUT/DELETE | Alert rules         | Yes           |
| `/api/alerts/incidents`      | GET/POST/PUT        | Incidents           | Yes           |
| `/api/settings/api-keys`     | GET/POST/DELETE     | API keys            | Yes           |
| `/api/settings/team`         | GET                 | Team members        | Yes           |
| `/api/settings/team/invite`  | POST                | Invite member       | Yes           |
| `/api/settings/integrations` | GET                 | Integrations        | Yes           |

### Performance Metrics

The API is designed to handle:

- 1000+ error reports per minute
- Sub-100ms response times for cached data
- 99.9% uptime availability
- Horizontal scaling capability

For production deployments, monitor these key metrics:

- Error ingestion rate
- Cache hit ratio
- Database query performance
- Memory and CPU usage
- Response time percentiles
