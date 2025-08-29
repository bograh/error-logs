# Health Monitor Backend API Implementation Guide

## Overview

This document outlines the complete backend API implementation required to support the Health Monitor dashboard. The frontend application expects these endpoints to provide comprehensive error monitoring, analytics, and system management capabilities.

## Base Configuration

### Base URL

```
https://your-domain.com/api
```

### Authentication

All endpoints (except health check) require API key authentication:

```http
X-API-Key: your-api-key-here
```

### Response Format

All responses should follow this consistent format:

**Success Response:**

```json
{
  "data": { ... },
  "status": "success"
}
```

**Error Response:**

```json
{
  "error": "Error description",
  "status": "error"
}
```

## Core Endpoints

### 1. Health Check

```http
GET /health
```

**Purpose:** API health status check  
**Authentication:** Not required  
**Response:**

```json
{
  "status": "ok",
  "timestamp": "2025-01-20T10:30:00Z"
}
```

---

## 2. Error Management Endpoints

### Error Logging

```http
POST /api/errors
```

**Purpose:** Create new error entries  
**Request Body:**

```json
{
  "level": "error",
  "message": "Database connection failed",
  "stack_trace": "Error: Connection timeout\n    at Database.connect...",
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

**Response:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-01-20T10:30:00Z",
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
    "first_seen": "2025-01-20T10:30:00Z",
    "last_seen": "2025-01-20T10:30:00Z",
    "created_at": "2025-01-20T10:30:00Z",
    "updated_at": "2025-01-20T10:30:00Z"
  },
  "status": "success"
}
```

### Get Errors List

```http
GET /api/errors
```

**Purpose:** Retrieve paginated error list with filtering  
**Query Parameters:**

- `limit` (integer, 1-100): Number of errors to return. Default: 50
- `offset` (integer): Number of errors to skip. Default: 0
- `level` (string): Filter by error level (error, warning, info, debug)
- `source` (string): Filter by error source
- `resolved` (boolean): Filter by resolution status
- `search` (string): Search in error messages

**Examples:**

```http
GET /api/errors?limit=20&offset=0
GET /api/errors?level=error&source=frontend
GET /api/errors?resolved=false&limit=10
```

**Response:**

```json
{
  "data": {
    "errors": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "timestamp": "2025-01-20T10:30:00Z",
        "level": "error",
        "message": "Database connection failed",
        "source": "backend",
        "resolved": false,
        "count": 5
      }
    ],
    "total": 150,
    "page": 1,
    "limit": 50
  },
  "status": "success"
}
```

### Get Error by ID

```http
GET /api/errors/{id}
```

**Purpose:** Get complete error details  
**Response:** Full error object with all fields

### Resolve Error

```http
PUT /api/errors/{id}/resolve
```

**Purpose:** Mark an error as resolved  
**Response:**

```json
{
  "data": {
    "status": "resolved"
  },
  "status": "success"
}
```

### Delete Error

```http
DELETE /api/errors/{id}
```

**Purpose:** Delete an error  
**Response:** 204 No Content

---

## 3. Analytics Endpoints

### Error Statistics

```http
GET /api/stats
```

**Purpose:** Get comprehensive error statistics  
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

### Error Trends

```http
GET /api/analytics/trends
```

**Purpose:** Get error trend data for charts  
**Query Parameters:**

- `period` (string): time, week, month, year
- `group_by` (string): hour, day, week, month

**Response:**

```json
{
  "data": {
    "period": "week",
    "data_points": [
      {
        "timestamp": "2025-01-14T00:00:00Z",
        "error_count": 45,
        "resolved_count": 32,
        "critical_count": 3
      },
      {
        "timestamp": "2025-01-15T00:00:00Z",
        "error_count": 52,
        "resolved_count": 38,
        "critical_count": 1
      }
    ]
  },
  "status": "success"
}
```

### Performance Metrics

```http
GET /api/analytics/performance
```

**Purpose:** System performance analytics  
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

## 4. Monitoring Endpoints

### Service Health

```http
GET /api/monitoring/services
```

**Purpose:** Get health status of all services  
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
        "last_checked": "2025-01-20T10:29:00Z",
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
        "last_checked": "2025-01-20T10:29:00Z"
      }
    ],
    "overall_health": "healthy"
  },
  "status": "success"
}
```

### System Metrics

```http
GET /api/monitoring/metrics
```

**Purpose:** Get system performance metrics  
**Query Parameters:**

- `timeframe` (string): 1h, 6h, 24h, 7d, 30d

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

### Uptime Data

```http
GET /api/monitoring/uptime
```

**Purpose:** Historical uptime data  
**Response:**

```json
{
  "data": {
    "current_uptime_hours": 720.5,
    "uptime_percent_24h": 100.0,
    "uptime_percent_7d": 99.8,
    "uptime_percent_30d": 99.95,
    "incidents_count": 2,
    "last_downtime": "2025-01-10T14:30:00Z"
  },
  "status": "success"
}
```

---

## 5. Alert Management Endpoints

### Alert Rules

```http
GET /api/alerts/rules
```

**Purpose:** Get all alert rules  
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
        "enabled": true,
        "notifications": ["email", "slack"],
        "last_triggered": "2025-01-15T10:30:00Z",
        "created_at": "2025-01-10T09:00:00Z"
      }
    ]
  },
  "status": "success"
}
```

```http
POST /api/alerts/rules
```

**Purpose:** Create new alert rule  
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

```http
PUT /api/alerts/rules/{id}
```

**Purpose:** Update alert rule

```http
DELETE /api/alerts/rules/{id}
```

**Purpose:** Delete alert rule

### Incidents

```http
GET /api/alerts/incidents
```

**Purpose:** Get incident list  
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
        "created_at": "2025-01-20T14:30:00Z",
        "updated_at": "2025-01-20T14:45:00Z",
        "assigned_to": "user-123",
        "description": "Multiple database connection timeouts detected"
      }
    ]
  },
  "status": "success"
}
```

```http
POST /api/alerts/incidents
```

**Purpose:** Create new incident

```http
PUT /api/alerts/incidents/{id}
```

**Purpose:** Update incident status

---

## 6. Settings & Configuration Endpoints

### API Keys Management

```http
GET /api/settings/api-keys
```

**Purpose:** Get list of API keys  
**Response:**

```json
{
  "data": {
    "api_keys": [
      {
        "id": "key-1",
        "name": "Production API Key",
        "key_preview": "sk_live_****7890",
        "permissions": ["read", "write"],
        "last_used": "2025-01-20T10:30:00Z",
        "created_at": "2025-01-10T09:00:00Z"
      }
    ]
  },
  "status": "success"
}
```

```http
POST /api/settings/api-keys
```

**Purpose:** Create new API key  
**Request Body:**

```json
{
  "name": "Development Key",
  "permissions": ["read"],
  "expires_at": "2025-12-31T23:59:59Z"
}
```

```http
DELETE /api/settings/api-keys/{id}
```

**Purpose:** Revoke API key

### Team Management

```http
GET /api/settings/team
```

**Purpose:** Get team members  
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
        "last_active": "2025-01-20T10:30:00Z",
        "created_at": "2025-01-10T09:00:00Z"
      }
    ]
  },
  "status": "success"
}
```

```http
POST /api/settings/team/invite
```

**Purpose:** Invite team member  
**Request Body:**

```json
{
  "email": "newuser@company.com",
  "role": "viewer"
}
```

### Integrations

```http
GET /api/settings/integrations
```

**Purpose:** Get integration status  
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
        "last_test": "2025-01-20T10:30:00Z"
      }
    ]
  },
  "status": "success"
}
```

---

## 7. Real-time Features

### WebSocket Endpoints

For real-time error streaming and live updates:

```javascript
// WebSocket connection
const ws = new WebSocket('wss://your-domain.com/ws?token=your-api-key');

// Message formats
{
  "type": "new_error",
  "data": { /* full error object */ }
}

{
  "type": "error_resolved",
  "data": { "id": "error-id" }
}

{
  "type": "service_status_change",
  "data": { "service": "database", "status": "unhealthy" }
}
```

---

## 8. Data Models

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

---

## 9. Error Aggregation & Fingerprinting

### Fingerprint Algorithm

Errors should be grouped using a fingerprint based on:

- Error message (normalized)
- Stack trace (first 3 lines, normalized)
- Source location

Example fingerprint generation:

```javascript
function generateFingerprint(error) {
  const message = normalizeMessage(error.message);
  const stackTrace = normalizeStackTrace(error.stack_trace);
  return hash(message + stackTrace + error.source);
}
```

### Aggregation Rules

- Increment `count` for duplicate fingerprints
- Update `last_seen` timestamp
- Keep `first_seen` unchanged
- Merge context data intelligently

---

## 10. Rate Limiting & Performance

### Rate Limits

- Error logging: 1000 requests/minute per API key
- Analytics queries: 100 requests/minute per API key
- General API: 500 requests/minute per API key

### Caching Strategy

- Error lists: Cache for 2 minutes
- Statistics: Cache for 1 minute
- Service health: Cache for 30 seconds
- Configuration: Cache for 10 minutes

### Database Indexes

Required indexes for performance:

```sql
-- Errors table
CREATE INDEX idx_errors_timestamp ON errors(timestamp);
CREATE INDEX idx_errors_level ON errors(level);
CREATE INDEX idx_errors_source ON errors(source);
CREATE INDEX idx_errors_resolved ON errors(resolved);
CREATE INDEX idx_errors_fingerprint ON errors(fingerprint);

-- Composite indexes
CREATE INDEX idx_errors_level_timestamp ON errors(level, timestamp);
CREATE INDEX idx_errors_resolved_timestamp ON errors(resolved, timestamp);
```

---

## 11. Security Requirements

### API Key Management

- Use secure random generation for API keys
- Store hashed versions in database
- Support key rotation without downtime
- Track key usage and last access

### Data Privacy

- Sanitize sensitive data in error contexts
- Support PII redaction
- Implement data retention policies
- Provide data export capabilities

### Access Control

Role-based permissions:

- **Owner**: Full access to all features
- **Admin**: Manage settings, view all data
- **Developer**: View errors, manage rules
- **Viewer**: Read-only access

---

## 12. Implementation Checklist

### Phase 1: Core Error Logging

- [ ] Error ingestion endpoint
- [ ] Error storage with fingerprinting
- [ ] Basic error retrieval
- [ ] Error resolution functionality

### Phase 2: Analytics & Monitoring

- [ ] Statistics aggregation
- [ ] Trend analysis endpoints
- [ ] Service health monitoring
- [ ] Performance metrics collection

### Phase 3: Advanced Features

- [ ] Alert rule engine
- [ ] Incident management
- [ ] Real-time WebSocket updates
- [ ] Team & permission management

### Phase 4: Enterprise Features

- [ ] Advanced integrations
- [ ] Custom dashboards
- [ ] Data export/import
- [ ] Audit logging

---

This API implementation guide provides the complete backend structure needed to support the Health Monitor frontend application. Each endpoint is designed to provide the data and functionality expected by the corresponding frontend components.
