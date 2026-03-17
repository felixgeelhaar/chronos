#!/bin/bash

# Stop development environment
# Usage: ./scripts/stop-dev.sh [--clean]

set -e

echo "🛑 Stopping Ascend development environment..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Navigate to project root
cd "$(dirname "$0")/.."

# Check for --clean flag
CLEAN=false
if [ "$1" = "--clean" ]; then
    CLEAN=true
fi

# Stop containers
echo -e "${YELLOW}Stopping containers...${NC}"
docker-compose down

if [ "$CLEAN" = true ]; then
    echo -e "${YELLOW}Removing volumes and data...${NC}"
    docker-compose down -v

    # Remove temporary files
    rm -rf api/tmp
    rm -rf api/storage/videos/*

    echo -e "${RED}⚠️  All data has been removed!${NC}"
else
    echo -e "${GREEN}✓ Containers stopped (data preserved)${NC}"
    echo "  To remove all data, run: ./scripts/stop-dev.sh --clean"
fi

echo ""
echo -e "${GREEN}✓ Development environment stopped${NC}"
