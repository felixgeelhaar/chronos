#!/bin/bash

# Reset development environment (clean restart)
# Usage: ./scripts/reset-dev.sh

set -e

echo "🔄 Resetting Ascend development environment..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Navigate to project root
cd "$(dirname "$0")/.."

# Confirm action
read -p "This will delete all data and restart fresh. Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 1
fi

# Stop and clean
echo -e "${YELLOW}Stopping containers and removing volumes...${NC}"
docker-compose down -v

# Remove build cache
echo -e "${YELLOW}Removing build cache...${NC}"
docker-compose build --no-cache

# Remove temporary files
echo -e "${YELLOW}Cleaning temporary files...${NC}"
rm -rf api/tmp
rm -rf api/storage/videos/*

# Start fresh
echo -e "${YELLOW}Starting fresh environment...${NC}"
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build -d

# Wait for services
echo -e "${YELLOW}Waiting for services...${NC}"
sleep 10

# Check health
echo -e "${YELLOW}Checking services...${NC}"

if docker-compose exec -T postgres pg_isready -U ascend > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
fi

if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is ready${NC}"
fi

if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ API is ready${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Environment reset complete!${NC}"
echo "  API: http://localhost:8080"
