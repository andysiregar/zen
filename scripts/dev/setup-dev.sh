#!/bin/bash

# Development Environment Setup Script
set -e

echo "🚀 Setting up Multi-Tenant SaaS Platform Development Environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed. Please install Docker Compose first.${NC}"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}⚠️  Go is not installed.${NC}"
    read -p "Do you want to install Go 1.21? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}📦 Installing Go...${NC}"
        ./scripts/dev/install-go.sh
        source ~/.bashrc
    else
        echo -e "${RED}❌ Go is required. Exiting.${NC}"
        exit 1
    fi
fi

# Check if Make is installed
if ! command -v make &> /dev/null; then
    echo -e "${YELLOW}📦 Installing make...${NC}"
    sudo apt-get update && sudo apt-get install -y make
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "${BLUE}📝 Creating .env file...${NC}"
    cp .env.example .env
    echo -e "${YELLOW}⚠️  Please review and update .env file with your configuration${NC}"
fi

# Start infrastructure services
echo -e "${BLUE}🐳 Starting infrastructure services...${NC}"
docker-compose -f deployments/docker/docker-compose.dev.yml up -d

# Wait for services to be ready
echo -e "${BLUE}⏳ Waiting for services to be ready...${NC}"
sleep 30

# Check if PostgreSQL is ready
echo -e "${BLUE}🗄️  Checking PostgreSQL connection...${NC}"
max_attempts=30
attempt=0
while ! docker exec saas-postgres pg_isready -U saas_user -d master_db > /dev/null 2>&1; do
    if [ $attempt -ge $max_attempts ]; then
        echo -e "${RED}❌ PostgreSQL is not ready after $max_attempts attempts${NC}"
        exit 1
    fi
    echo "Waiting for PostgreSQL... (attempt $((attempt + 1))/$max_attempts)"
    sleep 2
    ((attempt++))
done

# Check if Redis is ready
echo -e "${BLUE}📦 Checking Redis connection...${NC}"
max_attempts=30
attempt=0
while ! docker exec saas-redis redis-cli ping > /dev/null 2>&1; do
    if [ $attempt -ge $max_attempts ]; then
        echo -e "${RED}❌ Redis is not ready after $max_attempts attempts${NC}"
        exit 1
    fi
    echo "Waiting for Redis... (attempt $((attempt + 1))/$max_attempts)"
    sleep 2
    ((attempt++))
done

echo -e "${GREEN}✅ Development environment setup completed!${NC}"
echo ""
echo -e "${BLUE}📊 Services available at:${NC}"
echo -e "  - PostgreSQL: localhost:5432"
echo -e "  - Redis: localhost:6379"  
echo -e "  - RabbitMQ Management: http://localhost:15672 (saas_user/saas_password)"
echo -e "  - Prometheus: http://localhost:9090"
echo -e "  - Grafana: http://localhost:3000 (admin/admin123)"
echo -e "  - Jaeger: http://localhost:16686"
echo -e "  - MailHog: http://localhost:8025"
echo ""
echo -e "${GREEN}🚀 You can now start developing!${NC}"
echo -e "${BLUE}📚 Next steps:${NC}"
echo -e "  1. Review .env configuration"
echo -e "  2. Run 'make run-tenant-service' to start tenant management"
echo -e "  3. Run 'make run-auth-service' to start authentication"
echo -e "  4. Run 'make run-api-gateway' to start API gateway"
echo ""
echo -e "${YELLOW}📖 Run 'make help' to see all available commands${NC}"