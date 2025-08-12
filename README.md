# Multi-Tenant SaaS Platform

A modern multi-tenant SaaS platform built with Go microservices, featuring ticketing system, project management, and real-time chat support.

## 🏗️ Architecture

This platform implements a microservices architecture with database-per-tenant isolation for maximum security and scalability.

### Tech Stack
- **Backend**: Go with Gin framework
- **Frontend**: Next.js 14+ with TypeScript, Tailwind CSS, Zustand
- **Databases**: PostgreSQL (master + per-tenant), Redis (caching/sessions)
- **Real-time**: WebSockets with Gorilla WebSocket
- **Message Queue**: Redis Streams
- **ORM**: GORM for database operations
- **Container**: Docker with Kubernetes deployment
- **Monitoring**: OpenTelemetry, Prometheus, Grafana

## 🚀 Services

### Phase 1: Foundation Services
1. **API Gateway** - Request routing, rate limiting, authentication
2. **Tenant Management Service** - Tenant CRUD, database provisioning
3. **Database Management Service** - DB provisioning, migrations, backups
4. **Authentication Service** - Multi-tenant auth, JWT, RBAC

### Phase 2: Core Business Services
5. **Ticket Management Service** - Ticket CRUD, workflows, SLA tracking
6. **Project Management Service** - Projects, tasks, Kanban, time tracking
7. **Chat Service** - Real-time messaging, agent routing, file uploads
8. **Notification Service** - Email, push, in-app notifications

### Phase 3: Platform Services
9. **File Storage Service** - Upload/download, CDN, tenant isolation
10. **Integration Service** - Webhooks, third-party APIs, data sync
11. **Reporting Service** - Analytics, dashboards, custom reports
12. **Billing Service** - Subscriptions, usage tracking, payments

### Phase 4: Infrastructure Services
13. **Background Jobs Service** - Async processing, scheduled tasks
14. **Monitoring Service** - Metrics, logging, health checks, alerts

## 📁 Project Structure

```
saas-platform/
├── services/                    # Microservices
│   ├── api-gateway/            # API Gateway service
│   ├── tenant-management/      # Tenant Management service
│   ├── auth-service/           # Authentication service
│   └── ...                     # Other services
├── shared/                     # Shared packages
│   └── pkg/                    # Common utilities
│       ├── auth/               # JWT utilities
│       ├── database/           # Database connection
│       ├── middleware/         # Shared middleware
│       ├── redis/              # Redis client
│       └── utils/              # Common utilities
├── deployments/                # Deployment configurations
│   ├── docker/                 # Docker configurations
│   ├── k8s/                    # Kubernetes manifests
│   └── helm/                   # Helm charts
├── scripts/                    # Development scripts
│   ├── dev/                    # Development utilities
│   ├── deploy/                 # Deployment scripts
│   └── migration/              # Database migrations
├── docs/                       # Documentation
│   ├── api/                    # API documentation
│   ├── architecture/           # Architecture docs
│   └── development/            # Development guides
└── frontend/                   # Next.js frontend application
```

## 🛠️ Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Node.js 18+ (for frontend)
- Make

### Quick Setup

1. **Clone and setup environment**:
   ```bash
   git clone <repository>
   cd saas-platform
   cp .env.example .env
   ```

2. **Start development environment**:
   ```bash
   make setup
   ```

   This will:
   - Initialize Go modules for all services
   - Download dependencies
   - Start infrastructure services (PostgreSQL, Redis, etc.)
   - Run database migrations

3. **Start individual services** (in separate terminals):
   ```bash
   make run-tenant-service    # Port 8001
   make run-auth-service      # Port 8002
   make run-api-gateway       # Port 8000
   ```

### Development Commands

```bash
# Development Environment
make dev-up                    # Start all infrastructure services
make dev-down                  # Stop all infrastructure services
make dev-logs                  # Show logs
make dev-clean                 # Clean environment (removes data)

# Service Management
make init-services             # Initialize Go modules
make build                     # Build all services
make test                      # Run tests
make lint                      # Run linting
make deps                      # Download dependencies

# Database
make migrate-up                # Run migrations
make migrate-down              # Rollback migrations
make migrate-create name=xxx   # Create new migration

# Docker
make docker-build              # Build Docker images
make prod-build                # Build production images

# Development Shortcuts
make run-api-gateway          # Run API Gateway
make run-tenant-service       # Run Tenant Management
make run-auth-service         # Run Authentication
```

## 🌐 Development Services

When running `make dev-up`, the following services will be available:

| Service | URL | Credentials |
|---------|-----|-------------|
| PostgreSQL | localhost:5432 | saas_user/saas_password |
| Redis | localhost:6379 | - |
| RabbitMQ Management | http://localhost:15672 | saas_user/saas_password |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin123 |
| Jaeger | http://localhost:16686 | - |
| MailHog | http://localhost:8025 | - |

## 🔧 Configuration

### Environment Variables

Copy `.env.example` to `.env` and adjust the following key settings:

```bash
# Database
MASTER_DB_HOST=localhost
MASTER_DB_USER=saas_user
MASTER_DB_PASSWORD=saas_password

# JWT
JWT_SECRET=your-secret-key

# Services
API_GATEWAY_PORT=8000
TENANT_SERVICE_PORT=8001
```

### Service Ports

| Service | Development Port |
|---------|------------------|
| API Gateway | 8000 |
| Tenant Management | 8001 |
| Authentication | 8002 |
| Database Management | 8003 |
| Ticket Service | 8004 |
| Project Service | 8005 |
| Chat Service | 8006 |
| Notification Service | 8007 |

## 🏛️ Multi-Tenant Architecture

### Database Strategy
- **Master Database**: Stores tenant registry and global configuration
- **Tenant Databases**: Each tenant gets isolated database for security
- **Automatic Provisioning**: New tenant databases created automatically

### Tenant Resolution
- Subdomain-based routing (tenant1.yourdomain.com)
- Custom domain support
- Tenant context in all requests

## 📖 API Documentation

API documentation is available at:
- Swagger UI: http://localhost:8000/docs
- Redoc: http://localhost:8000/redoc

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests for specific service
cd services/tenant-management && go test ./...

# Run tests with coverage
cd services/tenant-management && go test -cover ./...
```

## 🚀 Deployment

### Docker
```bash
# Build images
make docker-build

# Production build
make prod-build
```

### Kubernetes
```bash
# Apply manifests
kubectl apply -f deployments/k8s/

# Using Helm
helm install saas-platform deployments/helm/
```

## 📊 Monitoring

The platform includes comprehensive monitoring:

- **Metrics**: Prometheus + Grafana dashboards
- **Tracing**: Jaeger for distributed tracing
- **Logging**: Structured logging with correlation IDs
- **Health Checks**: Service health endpoints

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- Documentation: `/docs`
- Issues: GitHub Issues
- Discussions: GitHub Discussions

---

**Next Steps**: Start with building the Tenant Management Service following the implementation guide in `/docs/development/`.