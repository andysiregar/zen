# CLAUDE.md

This file provides guidance for AI assistants when working with code in this repository.

## Project Overview

This is a Multi-Tenant SaaS Platform built with Go microservices, featuring ticketing system, project management, and real-time chat support. The platform implements a microservices architecture with database-per-tenant isolation for maximum security and scalability.

## Common Development Commands

### Environment Setup
```bash
# Complete setup from scratch
make setup

# Start infrastructure services only
make dev-up

# Stop infrastructure services
make dev-down

# View infrastructure logs
make dev-logs

# Clean development environment (removes all data)
make dev-clean
```

### Service Management
```bash
# Initialize Go modules for all services
make init-services

# Download dependencies for all services
make deps

# Build all services
make build

# Run tests for all services
make test

# Run linting for all services (uses golangci-lint)
make lint

# Format all Go code
make format

# Clean build artifacts
make clean
```

### Individual Service Development
```bash
# Run individual services
make run-api-gateway          # Port 8000
make run-tenant-service       # Port 8001
make run-auth-service         # Port 8002
make run-platform-admin       # Port 8014

# Build specific service
cd services/[service-name]
go build -o bin/[service-name] cmd/main.go

# Test specific service
cd services/[service-name]
go test ./...

# Test with coverage
cd services/[service-name]
go test -cover ./...
```

### Database Management
```bash
# Run migrations up
make migrate-up

# Run migrations down  
make migrate-down

# Create new migration
make migrate-create name=migration_name
```

### Docker Operations
```bash
# Build Docker images for all services
make docker-build

# Build production Docker images
make prod-build
```

## Architecture

### Tech Stack
- **Backend**: Go with Gin framework
- **Frontend**: Next.js 14+ with TypeScript, Tailwind CSS, Zustand
- **Databases**: PostgreSQL (master + per-tenant), Redis (caching/sessions)
- **Real-time**: WebSockets with Gorilla WebSocket
- **Message Queue**: Redis Streams
- **ORM**: GORM for database operations
- **Container**: Docker with Kubernetes deployment
- **Monitoring**: OpenTelemetry, Prometheus, Grafana

### Service Architecture
The platform consists of 14 microservices organized in 4 phases:

**Phase 1 - Foundation Services:**
1. API Gateway (Port 8000) - Request routing, rate limiting, authentication
2. Tenant Management (Port 8001) - Tenant CRUD, database provisioning
3. Database Management (Port 8003) - DB provisioning, migrations, backups
4. Authentication Service (Port 8002) - Multi-tenant auth, JWT, RBAC

**Phase 2 - Core Business Services:**
5. Ticket Service (Port 8004) - Ticket CRUD, workflows, SLA tracking
6. Project Service (Port 8005) - Projects, tasks, Kanban, time tracking
7. Chat Service (Port 8006) - Real-time messaging, agent routing
8. Notification Service (Port 8007) - Email, push, in-app notifications

**Phase 3 - Platform Services:**
9. File Storage Service - Upload/download, CDN, tenant isolation
10. Integration Service - Webhooks, third-party APIs, data sync
11. Reporting Service - Analytics, dashboards, custom reports
12. Billing Service - Subscriptions, usage tracking, payments

**Phase 4 - Infrastructure Services:**
13. Background Jobs Service - Async processing, scheduled tasks
14. Monitoring Service - Metrics, logging, health checks, alerts

### Multi-Tenant Architecture
- **Master Database**: Stores tenant registry and global configuration
- **Tenant Databases**: Each tenant gets isolated database for security
- **Automatic Provisioning**: New tenant databases created automatically
- **Tenant Resolution**: Subdomain-based routing (tenant1.yourdomain.com)

### Directory Structure
- `services/` - Individual microservices with standard Go project layout
- `shared/pkg/` - Common utilities shared across services:
  - `auth/` - JWT utilities and authentication logic
  - `database/` - Database connection and configuration
  - `middleware/` - Shared middleware (CORS, logging, etc.)
  - `redis/` - Redis client utilities
  - `utils/` - Common response and validation utilities
- `deployments/` - Docker, Kubernetes, and Helm configurations
- `scripts/` - Development and deployment scripts
- `docs/` - API documentation, architecture docs, development guides

### Service Structure Pattern
Each service follows a consistent internal structure:
```
services/[service-name]/
├── cmd/main.go           # Application entry point
├── internal/
│   ├── config/           # Service configuration
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Service-specific middleware
│   ├── models/           # Data models
│   ├── repositories/     # Data access layer
│   ├── services/         # Business logic
│   └── utils/            # Service utilities
├── pkg/                  # Public API/packages
├── tests/                # Test files
├── migrations/           # Database migrations
├── docker/               # Docker configurations
└── k8s/                  # Kubernetes manifests
```

## Development Environment Services

When running `make dev-up`, these services become available:

| Service | URL | Credentials |
|---------|-----|-------------|
| PostgreSQL | localhost:5432 | saas_user/saas_password |
| Redis | localhost:6379 | - |
| RabbitMQ Management | http://localhost:15672 | saas_user/saas_password |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin123 |
| Jaeger | http://localhost:16686 | - |
| MailHog | http://localhost:8025 | - |

## Testing

### Testing Commands
- `make test` - Run tests for all services
- `cd services/[service] && go test ./...` - Test specific service
- `cd services/[service] && go test -cover ./...` - Test with coverage
- `cd services/[service] && go test ./internal/services -v` - Unit tests
- `cd services/[service] && go test ./tests/integration -v` - Integration tests

### Database Connections for Testing
```bash
# Connect to master database
docker exec -it saas-postgres psql -U saas_user -d master_db

# Connect to tenant database
docker exec -it saas-postgres psql -U saas_user -d tenant_demo
```

## Key Dependencies and Frameworks

All services use these common dependencies:
- `github.com/gin-gonic/gin` - HTTP web framework
- `gorm.io/gorm` and `gorm.io/driver/postgres` - ORM and PostgreSQL driver
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `github.com/redis/go-redis/v9` - Redis client
- `go.uber.org/zap` - Structured logging
- `github.com/gin-contrib/cors` - CORS middleware
- `github.com/prometheus/client_golang` - Prometheus metrics
- `go.opentelemetry.io/otel` - OpenTelemetry tracing
- `github.com/gorilla/websocket` - WebSocket support

## Configuration

Services use environment variables for configuration. Key variables include:
- `MASTER_DB_HOST`, `MASTER_DB_USER`, `MASTER_DB_PASSWORD` - Master database connection
- `REDIS_HOST`, `REDIS_PORT` - Redis connection
- `JWT_SECRET` - JWT signing secret
- `API_GATEWAY_PORT`, `TENANT_SERVICE_PORT`, etc. - Service ports

## API Documentation

API documentation is available at:
- Swagger UI: http://localhost:8000/docs
- Redoc: http://localhost:8000/redoc

## Current Development Status (Last Updated: 2025-08-12)

### ✅ Completed Foundation Services
- **Go Installation**: Go 1.21.5 installed at `/usr/local/go/bin`
- **Core Foundation Services** (All building successfully):
  - **API Gateway** (`services/api-gateway/cmd/main.go`) - Port 8000 ✅
  - **Tenant Management** (`services/tenant-management/cmd/main.go`) - Port 8001 ✅
  - **Authentication Service** (`services/auth-service/cmd/main.go`) - Port 8002 ✅
  - **Platform Admin Service** (`services/platform-admin/cmd/main.go`) - Port 8014 ✅
- **Shared Utilities**: Complete JWT, database, Redis, middleware, utils in `shared/pkg/`
- **Database Models**: User, Tenant, Organization, PlatformAdmin models with GORM
- **RBAC System**: Complete platform admin role-based access control

### 🚧 Foundation Services Implementation Status
Each foundation service has:
- ✅ HTTP server with Gin framework
- ✅ Graceful shutdown handling  
- ✅ Health check endpoints (`/health`)
- ✅ CORS middleware
- ✅ Structured logging with Zap
- ✅ Configuration management (internal/config/)
- ✅ Database models and repositories
- ✅ Business logic implementation (internal/services/)
- ✅ Complete HTTP handlers with real functionality
- ✅ Multi-tenant database connection management
- ✅ JWT-based authentication with role support

### ❌ Still Missing
- Database migrations execution
- Remaining 10 services (database-management, ticket-service, project-service, etc.)
- Infrastructure setup (PostgreSQL, Redis containers)
- Integration testing between services

### 🎯 Immediate Next Steps
When resuming development, continue with:
1. **Set up infrastructure**: `make dev-up` to start PostgreSQL, Redis, etc.
2. **Run database migrations**: `make migrate-up` to create all database tables
3. **Test foundation services**: Start services and test authentication flow
4. **Build Phase 2 services**: ticket-service, project-service, chat-service, notification-service
5. **API Gateway integration**: Configure routing to all foundation services
6. **End-to-end testing**: Complete multi-tenant authentication and tenant management flow

### 🔧 Build Commands (Remember to export Go PATH)
```bash
export PATH="/usr/local/go/bin:$PATH"
cd services/[service-name]
go mod tidy
go build -o bin/[service-name] cmd/main.go
```

## Git Configuration

When making commits, AI assistants should:
- Use the repository owner's name and email for commits
- Do not include AI-generated commit footers or co-authoring tags
- Focus on clear, descriptive commit messages that explain the changes made