# Multi-Tenant SaaS Platform Makefile

# Variables
DOCKER_COMPOSE_DEV = deployments/docker/docker-compose.dev.yml
SERVICES := api-gateway tenant-management database-management auth-service ticket-service project-service chat-service notification-service file-storage integration-service reporting-service billing-service background-jobs monitoring-service

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

.PHONY: help dev-up dev-down dev-logs build test lint clean init-services migrate-up migrate-down

help: ## Show this help message
	@echo "$(BLUE)Multi-Tenant SaaS Platform Development Commands$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

# Development Environment
dev-up: ## Start development environment
	@echo "$(YELLOW)Starting development environment...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_DEV) up -d
	@echo "$(GREEN)Development environment started!$(RESET)"
	@echo "$(BLUE)Services available at:$(RESET)"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - RabbitMQ: localhost:15672 (admin/admin)"
	@echo "  - Prometheus: localhost:9090"
	@echo "  - Grafana: localhost:3000 (admin/admin123)"
	@echo "  - Jaeger: localhost:16686"
	@echo "  - MailHog: localhost:8025"

dev-down: ## Stop development environment
	@echo "$(YELLOW)Stopping development environment...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_DEV) down
	@echo "$(GREEN)Development environment stopped!$(RESET)"

dev-logs: ## Show development environment logs
	docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f

dev-clean: ## Clean development environment (remove volumes)
	@echo "$(RED)Warning: This will delete all data in the development environment$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose -f $(DOCKER_COMPOSE_DEV) down -v; \
		echo "$(GREEN)Development environment cleaned!$(RESET)"; \
	else \
		echo "$(YELLOW)Cancelled$(RESET)"; \
	fi

# Service Management
init-services: ## Initialize Go modules for all services
	@echo "$(YELLOW)Initializing Go modules for all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Initializing $$service...$(RESET)"; \
		cd services/$$service && go mod init $$service && cd ../..; \
	done
	@echo "$(GREEN)All services initialized!$(RESET)"

build: ## Build all services
	@echo "$(YELLOW)Building all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Building $$service...$(RESET)"; \
		cd services/$$service && go build -o bin/$$service cmd/main.go && cd ../..; \
	done
	@echo "$(GREEN)All services built!$(RESET)"

test: ## Run tests for all services
	@echo "$(YELLOW)Running tests for all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Testing $$service...$(RESET)"; \
		cd services/$$service && go test ./... && cd ../..; \
	done
	@echo "$(GREEN)All tests completed!$(RESET)"

lint: ## Run linting for all services
	@echo "$(YELLOW)Running linting for all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Linting $$service...$(RESET)"; \
		cd services/$$service && golangci-lint run && cd ../..; \
	done
	@echo "$(GREEN)All linting completed!$(RESET)"

# Database Management
migrate-up: ## Run database migrations up
	@echo "$(YELLOW)Running database migrations up...$(RESET)"
	cd scripts/migration && go run migrate.go up
	@echo "$(GREEN)Migrations completed!$(RESET)"

migrate-down: ## Run database migrations down
	@echo "$(YELLOW)Running database migrations down...$(RESET)"
	cd scripts/migration && go run migrate.go down
	@echo "$(GREEN)Migrations rolled back!$(RESET)"

migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Please provide a migration name. Usage: make migrate-create name=migration_name$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Creating migration: $(name)...$(RESET)"
	cd scripts/migration && go run migrate.go create $(name)
	@echo "$(GREEN)Migration created!$(RESET)"

# Development Tools
format: ## Format all Go code
	@echo "$(YELLOW)Formatting all Go code...$(RESET)"
	@find . -name "*.go" -exec gofmt -w {} \;
	@echo "$(GREEN)All code formatted!$(RESET)"

deps: ## Download dependencies for all services
	@echo "$(YELLOW)Downloading dependencies for all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Downloading deps for $$service...$(RESET)"; \
		cd services/$$service && go mod tidy && cd ../..; \
	done
	@echo "$(GREEN)All dependencies downloaded!$(RESET)"

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "*.log" -delete 2>/dev/null || true
	@echo "$(GREEN)Cleanup completed!$(RESET)"

# Docker Management
docker-build: ## Build all Docker images
	@echo "$(YELLOW)Building all Docker images...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Building Docker image for $$service...$(RESET)"; \
		cd services/$$service && docker build -t $$service:latest -f docker/Dockerfile . && cd ../..; \
	done
	@echo "$(GREEN)All Docker images built!$(RESET)"

# Development Shortcuts
run-api-gateway: ## Run API Gateway service
	cd services/api-gateway && go run cmd/main.go

run-tenant-service: ## Run Tenant Management service
	cd services/tenant-management && go run cmd/main.go

run-auth-service: ## Run Authentication service
	cd services/auth-service && go run cmd/main.go

# Setup Commands
setup: ## Setup development environment from scratch
	@echo "$(YELLOW)Setting up development environment...$(RESET)"
	@make init-services
	@make deps
	@make dev-up
	@sleep 10  # Wait for services to start
	@make migrate-up
	@echo "$(GREEN)Development environment setup completed!$(RESET)"
	@echo "$(BLUE)You can now start developing!$(RESET)"

# Production
prod-build: ## Build production Docker images
	@echo "$(YELLOW)Building production Docker images...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Building production image for $$service...$(RESET)"; \
		cd services/$$service && docker build -t $$service:prod -f docker/Dockerfile.prod . && cd ../..; \
	done
	@echo "$(GREEN)All production images built!$(RESET)"

# Git Branch Management
sync-branches: ## Sync develop and main branches intelligently
	@echo "$(YELLOW)Running branch synchronization script...$(RESET)"
	@bash scripts/dev/sync-branches.sh