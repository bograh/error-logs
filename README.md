# Minimal Sentry Clone - Error Tracking System

A lightweight, open-source error tracking platform built with **Go**, **Next.js**, **PostgreSQL**, and **Redis**. This system captures, stores, and displays application errors in real-time with a clean dashboard interface.

## 🏗️ Architecture Overview

```
Frontend (Next.js)          Backend (Go + Chi)         Storage
     │                           │                       │
     ├─ Dashboard UI             ├─ REST API             ├─ PostgreSQL
     ├─ Error Details            ├─ API Key Auth         │  (Persistent Storage)
     ├─ Test Error Buttons       ├─ Error Processing     │
     └─ Real-time Updates        └─ Background Workers   └─ Redis
                                                          (Caching & Queuing)
```

### Data Flow:

1. **Error Capture**: Frontend/Backend sends error → Go API
2. **Immediate Storage**: Error queued in Redis for fast response
3. **Background Processing**: Worker processes Redis queue → PostgreSQL
4. **Caching**: Frequently accessed data cached in Redis with TTL
5. **Display**: Dashboard fetches from cache/database → User interface

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for local development)
- Node.js 18+ (for local development)

### 1. Clone and Setup

```bash
git clone <your-repo>
cd error-logs
```

### 2. Start with Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

This will start:

- **PostgreSQL** on port 5432
- **Redis** on port 6379
- **Go Backend** on port 8080
- **Next.js Frontend** on port 3000

### 3. Access the Application

- **Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### 4. Test the System

1. Visit the dashboard at http://localhost:3000
2. Click "Trigger Frontend Error" or "Trigger Backend Error"
3. Watch errors appear in real-time on the dashboard
4. Click on errors to view detailed information

## 🛠️ Local Development Setup

### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Set up environment
cp .env.example .env

# Start PostgreSQL and Redis
docker-compose up postgres redis -d

# Run the backend
go run main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Set up environment
cp .env.local.example .env.local

# Start development server
npm run dev
```

## 📊 Database Schema

### Main Tables:

- **`errors`**: Core error events with metadata
- **`api_keys`**: Authentication keys for API access
- **`projects`**: Multi-project support (future enhancement)

### Key Fields in `errors` table:

- `id` (UUID): Unique identifier
- `timestamp`: When error occurred
- `level`: error, warning, info, debug
- `message`: Error message
- `stack_trace`: Full stack trace
- `context`: JSON metadata
- `source`: frontend, backend, api
- `fingerprint`: For grouping similar errors
- `resolved`: Resolution status

## 🔌 API Endpoints

### Authentication

All API requests require an `X-API-Key` header:

```bash
X-API-Key: test-api-key
```

### Core Endpoints:

```bash
# Create Error
POST /api/errors
{
  "level": "error",
  "message": "Something went wrong",
  "stack_trace": "Error: ...",
  "context": {"user_id": 123},
  "source": "frontend",
  "url": "https://app.com/page"
}

# List Errors
GET /api/errors?limit=50&offset=0&level=error&source=frontend

# Get Error Details
GET /api/errors/{id}

# Resolve Error
PUT /api/errors/{id}/resolve

# Delete Error
DELETE /api/errors/{id}

# Get Statistics
GET /api/stats
```

## 🧪 Testing & Error Flow

### Manual Testing:

1. Use the "Trigger Error" buttons in the dashboard
2. Monitor error appearance in the UI
3. Test error resolution and deletion

### API Testing with curl:

```bash
# Create a test error
curl -X POST http://localhost:8080/api/errors \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "level": "error",
    "message": "Test API error",
    "source": "api",
    "context": {"test": true}
  }'

# Get errors list
curl -H "X-API-Key: test-api-key" \
  "http://localhost:8080/api/errors?limit=5"
```

### Integration Test Flow:

1. **Create Error** → Verify Redis queuing
2. **Background Processing** → Verify PostgreSQL persistence
3. **Caching** → Verify Redis cache population
4. **API Retrieval** → Verify data accuracy
5. **Frontend Display** → Verify UI updates

## 🔧 Configuration

### Environment Variables:

#### Backend (.env):

```bash
DATABASE_URL=postgres://error_logs_user:error_logs_password@localhost:5432/error_logs?sslmode=disable
REDIS_URL=redis://localhost:6379
PORT=8080
ENVIRONMENT=development
```

#### Frontend (.env.local):

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NODE_ENV=development
```

## 🎯 Key Features Implemented

### ✅ Core Functionality:

- [x] Error capture and storage
- [x] Real-time dashboard
- [x] Error details view
- [x] Redis caching and queuing
- [x] Background processing
- [x] API key authentication
- [x] Statistics tracking
- [x] Docker deployment

### ✅ UI Components:

- [x] Modern dashboard with Shadcn/ui
- [x] Error list with filtering
- [x] Detailed error modal
- [x] Statistics cards
- [x] Test error triggers
- [x] Responsive design

### ✅ Backend Features:

- [x] Chi router with middleware
- [x] PostgreSQL with optimized schema
- [x] Redis integration
- [x] Background queue processing
- [x] API key validation
- [x] Error fingerprinting
- [x] CORS support

## 🚀 Deployment Options

### Production Deployment with Docker:

```bash
# Build and deploy
docker-compose -f docker-compose.prod.yml up -d

# Scale backend if needed
docker-compose up --scale backend=3
```

### VPS Deployment:

1. Copy files to server
2. Set production environment variables
3. Run `docker-compose up -d`
4. Set up reverse proxy (nginx/traefik)
5. Configure SSL certificates

## 🔮 Optional Enhancements

### Immediate Improvements:

- [ ] **Error Grouping**: Group similar errors by fingerprint
- [ ] **Pagination**: Implement proper pagination in frontend
- [ ] **Real-time Updates**: WebSocket for live error notifications
- [ ] **Email Alerts**: Notify on critical errors
- [ ] **Rate Limiting**: Prevent spam/DoS

### Advanced Features:

- [ ] **User Authentication**: Multi-user support
- [ ] **Project Management**: Isolate errors by project
- [ ] **Advanced Filtering**: Date ranges, custom queries
- [ ] **Webhooks**: Integration with Slack/Discord
- [ ] **Performance Monitoring**: Track response times
- [ ] **Source Maps**: Better stack trace resolution
- [ ] **Error Trends**: Analytics and trending

### Scalability:

- [ ] **Horizontal Scaling**: Load balancer setup
- [ ] **Database Sharding**: Handle high volume
- [ ] **CDC (Change Data Capture)**: Real-time sync
- [ ] **Metrics**: Prometheus/Grafana monitoring
- [ ] **Logging**: Structured logging with ELK stack

## 🐛 Troubleshooting

### Common Issues:

1. **Connection Refused**:

   ```bash
   # Check if services are running
   docker-compose ps

   # Check logs
   docker-compose logs backend
   ```

2. **Database Connection Issues**:

   ```bash
   # Verify PostgreSQL is ready
   docker-compose exec postgres pg_isready -U error_logs_user
   ```

3. **Redis Connection Issues**:

   ```bash
   # Test Redis connectivity
   docker-compose exec redis redis-cli ping
   ```

4. **API Key Issues**:
   - Verify `X-API-Key: test-api-key` header
   - Check API key exists in database

## 📝 License

This project is open-source and available under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

**Built with ❤️ using Go, Next.js, PostgreSQL, and Redis**
