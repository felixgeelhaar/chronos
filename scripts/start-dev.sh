#!/bin/bash

# Start development environment with docker-compose
# Usage: ./scripts/start-dev.sh

set -e

echo "🚀 Starting Ascend development environment..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Navigate to project root
cd "$(dirname "$0")/.."

# Stop any running containers
echo -e "${YELLOW}Stopping existing containers...${NC}"
docker-compose down

# Build and start services
echo -e "${YELLOW}Building and starting services...${NC}"
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build -d

# Wait for services to be healthy
echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
sleep 5

# Check service health
echo -e "${YELLOW}Checking service health...${NC}"

# Check PostgreSQL
if docker-compose exec -T postgres pg_isready -U ascend > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
else
    echo "⚠️  PostgreSQL is not ready yet"
fi

# Check Redis
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is ready${NC}"
else
    echo "⚠️  Redis is not ready yet"
fi

# Check API
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ API is ready${NC}"
else
    echo "⚠️  API is not ready yet (might still be starting)"
fi

echo ""
echo -e "${GREEN}🎉 Development environment is running!${NC}"
echo ""
echo "Services:"
echo "  API:      http://localhost:8080"
echo "  Adminer:  http://localhost:8081 (Database UI)"
echo "  MinIO:    http://localhost:9001 (Object Storage UI)"
echo ""
echo "Database:"
echo "  Host:     localhost"
echo "  Port:     5432"
echo "  Database: ascend"
echo "  User:     ascend"
echo "  Password: ascend_dev_password"
echo ""
echo "Commands:"
echo "  View logs:    docker-compose logs -f"
echo "  View API logs: docker-compose logs -f api"
echo "  Stop:         ./scripts/stop-dev.sh"
echo "  Restart:      docker-compose restart api"
echo ""
