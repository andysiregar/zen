# Getting Started with Development

This guide will help you set up the development environment for the Multi-Tenant SaaS Platform.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Docker & Docker Compose**: For running infrastructure services
- **Go 1.21+**: For backend development
- **Node.js 18+**: For frontend development (optional for backend-only development)
- **Make**: For running build commands

## Quick Setup

### 1. Clone and Setup

```bash
git clone <repository-url>
cd saas-platform
```

### 2. Automated Setup

Run the automated setup script:

```bash
./scripts/dev/setup-dev.sh
```

This script will:
- Install Go if not present
- Create `.env` file from template
- Start all infrastructure services
- Verify service connections

### 3. Manual Setup (Alternative)

If you prefer manual setup:

```bash
# Copy environment template
cp .env.example .env

# Start infrastructure services
make dev-up

# Initialize services (if Go is available)
make init-services
make deps
```

## Infrastructure Services

After running setup, the following services will be available:

| Service | URL | Credentials |
|---------|-----|-------------|
| PostgreSQL | localhost:5432 | saas_user/saas_password |
| Redis | localhost:6379 | - |
| RabbitMQ Management | http://localhost:15672 | saas_user/saas_password |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin123 |
| Jaeger | http://localhost:16686 | - |
| MailHog | http://localhost:8025 | - |

## Starting Services

### Individual Services

Start each microservice in separate terminals:

```bash
# Terminal 1: Tenant Management Service
make run-tenant-service

# Terminal 2: Authentication Service  
make run-auth-service

# Terminal 3: API Gateway
make run-api-gateway
```

### Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| API Gateway | 8000 | Main entry point |
| Tenant Management | 8001 | Tenant operations |
| Authentication | 8002 | User auth |
| Database Management | 8003 | DB operations |
| Ticket Service | 8004 | Ticketing system |
| Project Service | 8005 | Project management |
| Chat Service | 8006 | Real-time chat |
| Notification Service | 8007 | Notifications |

## Development Workflow

### 1. Code Changes

Make your changes to the relevant service in `services/<service-name>/`

### 2. Testing

```bash
# Test specific service
cd services/tenant-management
go test ./...

# Test all services
make test
```

### 3. Building

```bash
# Build specific service
cd services/tenant-management
go build -o bin/tenant-management cmd/main.go

# Build all services
make build
```

### 4. Linting

```bash
# Lint all services
make lint
```

## Database Management

### Migrations

Create new migration:
```bash
make migrate-create name=create_users_table
```

Run migrations:
```bash
make migrate-up
```

Rollback migrations:
```bash
make migrate-down
```

### Connecting to Database

```bash
# Connect to master database
docker exec -it saas-postgres psql -U saas_user -d master_db

# Connect to tenant database
docker exec -it saas-postgres psql -U saas_user -d tenant_demo
```

## Configuration

### Environment Variables

Key environment variables in `.env`:

```bash
# Database
MASTER_DB_HOST=localhost
MASTER_DB_USER=saas_user
MASTER_DB_PASSWORD=saas_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key

# Service Ports
API_GATEWAY_PORT=8000
TENANT_SERVICE_PORT=8001
```

### Service Configuration

Each service has its own configuration in `internal/config/config.go`:

```go
type Config struct {
    ServiceName string
    Port        string
    DBConfig    database.Config
    RedisConfig redis.Config
    JWTSecret   string
}
```

## Debugging

### Logs

View service logs:
```bash
# Infrastructure services
make dev-logs

# Individual service logs
cd services/tenant-management
go run cmd/main.go
```

### Debugging with VS Code

Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Tenant Service",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/services/tenant-management/cmd/main.go",
            "env": {
                "GIN_MODE": "debug",
                "SERVICE_PORT": "8001"
            }
        }
    ]
}
```

## Testing

### Unit Tests

```bash
cd services/tenant-management
go test ./internal/services -v
```

### Integration Tests

```bash
cd services/tenant-management
go test ./tests/integration -v
```

### API Testing

Use the provided Postman collection or curl:

```bash
# Health check
curl http://localhost:8000/health

# Create tenant
curl -X POST http://localhost:8001/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"subdomain":"test","plan_type":"basic"}'
```

## Common Issues

### 1. Port Already in Use

```bash
# Find process using port
lsof -i :8001

# Kill process
kill -9 <PID>
```

### 2. Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Restart PostgreSQL
docker restart saas-postgres
```

### 3. Go Module Issues

```bash
# Clean Go modules
go clean -modcache
go mod download
```

## Next Steps

1. **Build your first service**: Start with [Tenant Management Service](./tenant-management.md)
2. **Understanding the architecture**: Read [Architecture Overview](../architecture/overview.md)
3. **API Development**: Follow [API Guidelines](./api-guidelines.md)

## Getting Help

- Check the [troubleshooting guide](./troubleshooting.md)
- Review service-specific documentation in `docs/`
- Open an issue for bugs or feature requests