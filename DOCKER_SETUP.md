# Docker Setup Guide

Complete Docker setup for the Ascend application, including backend API, database, cache, and object storage.

## Prerequisites

- Docker Desktop (or Docker Engine + Docker Compose)
- At least 4GB RAM available for Docker
- Port availability: 5432 (PostgreSQL), 6379 (Redis), 8080 (API), 8081 (Adminer), 9000-9001 (MinIO)

## Quick Start

### Start Development Environment

```bash
./scripts/start-dev.sh
```

This will:
- Build all Docker images
- Start PostgreSQL, Redis, MinIO, API, and Adminer
- Initialize the database with migrations
- Enable hot-reload for the API

### Stop Development Environment

```bash
# Stop containers (preserve data)
./scripts/stop-dev.sh

# Stop and remove all data
./scripts/stop-dev.sh --clean
```

### Reset Environment

```bash
# Complete reset (removes all data and rebuilds)
./scripts/reset-dev.sh
```

## Services

### API Server (Port 8080)
- **URL:** http://localhost:8080
- **Health Check:** http://localhost:8080/health
- **Features:**
  - Hot-reload with Air (development)
  - Automatic restart on code changes
  - Request logging
  - CORS enabled for mobile development

### PostgreSQL Database (Port 5432)
- **Host:** localhost
- **Port:** 5432
- **Database:** ascend
- **User:** ascend
- **Password:** ascend_dev_password
- **Features:**
  - Persistent data storage
  - Auto-initialization with migrations
  - Health checks

### Redis Cache (Port 6379)
- **Host:** localhost
- **Port:** 6379
- **Features:**
  - Session storage
  - Caching layer
  - Pub/sub messaging

### Adminer - Database UI (Port 8081)
- **URL:** http://localhost:8081
- **System:** PostgreSQL
- **Server:** postgres
- **Username:** ascend
- **Password:** ascend_dev_password
- **Features:**
  - Web-based SQL client
  - Schema visualization
  - Query execution
  - Data export/import

### MinIO - Object Storage (Ports 9000, 9001)
- **API:** http://localhost:9000
- **Console:** http://localhost:9001
- **Username:** ascend
- **Password:** ascend_minio_password
- **Features:**
  - S3-compatible API
  - Video file storage
  - Web console for file management

## Configuration

### Environment Variables

The API service uses the following environment variables (configured in docker-compose.yml):

#### Database
```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=ascend
DB_PASSWORD=ascend_dev_password
DB_NAME=ascend
DB_SSL_MODE=disable
```

#### Redis
```env
REDIS_HOST=redis
REDIS_PORT=6379
```

#### Server
```env
PORT=8080
ENV=development
```

#### JWT
```env
JWT_SECRET=dev_jwt_secret_change_in_production
JWT_EXPIRATION=24h
```

#### Video Storage
```env
VIDEO_STORAGE_PATH=/app/storage/videos
MAX_VIDEO_SIZE_MB=100
ALLOWED_VIDEO_FORMATS=mp4,mov,avi
```

#### CORS
```env
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006
```

### Customizing Configuration

Create a `.env` file in the project root to override default values:

```env
# .env
API_PORT=8080
POSTGRES_PASSWORD=custom_password
REDIS_PORT=6379
```

Then reference it in docker-compose.yml:

```yaml
services:
  api:
    env_file:
      - .env
```

## Docker Commands

### Basic Operations

```bash
# Start all services
docker-compose up -d

# Start with development overrides
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Rebuild images
docker-compose build --no-cache

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api
docker-compose logs -f postgres
```

### Service Management

```bash
# Restart a service
docker-compose restart api

# Stop a service
docker-compose stop api

# Start a service
docker-compose start api

# Scale a service (if applicable)
docker-compose up -d --scale api=3
```

### Debugging

```bash
# Execute command in running container
docker-compose exec api sh

# Run one-off command
docker-compose run --rm api go version

# Check service health
docker-compose ps

# Inspect a service
docker-compose exec api env

# Database shell
docker-compose exec postgres psql -U ascend -d ascend

# Redis CLI
docker-compose exec redis redis-cli
```

## Development Workflow

### 1. Start Environment
```bash
./scripts/start-dev.sh
```

### 2. Develop with Hot-Reload
- Edit code in `api/` directory
- Changes automatically trigger rebuild
- API restarts automatically
- View logs: `docker-compose logs -f api`

### 3. Test API Endpoints
```bash
# Health check
curl http://localhost:8080/health

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Get sessions
curl http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <token>"
```

### 4. Database Operations

**Via Adminer (Web UI):**
1. Open http://localhost:8081
2. Login with credentials
3. Browse tables, run queries, view data

**Via psql (CLI):**
```bash
# Connect to database
docker-compose exec postgres psql -U ascend -d ascend

# Run queries
SELECT * FROM users;
SELECT * FROM sessions;
```

### 5. Monitor Services
```bash
# View all logs
docker-compose logs -f

# View API logs only
docker-compose logs -f api

# View last 100 lines
docker-compose logs --tail=100 api
```

## Mobile App Integration

### Connect Mobile App to Docker Backend

Update mobile app `.env`:

```env
# For iOS Simulator
API_BASE_URL=http://localhost:8080

# For Android Emulator
API_BASE_URL=http://10.0.2.2:8080

# For Physical Device (use your computer's IP)
API_BASE_URL=http://192.168.1.XXX:8080
```

### Enable CORS for Mobile Development

The docker-compose configuration already includes:

```yaml
ALLOWED_ORIGINS: http://localhost:3000,http://localhost:19006
```

For physical devices, add your IP:

```yaml
ALLOWED_ORIGINS: http://localhost:3000,http://192.168.1.XXX:19006
```

## Production Deployment

### Build Production Images

```bash
# Build production image
docker build -t ascend-api:latest ./api

# Tag for registry
docker tag ascend-api:latest registry.example.com/ascend-api:latest

# Push to registry
docker push registry.example.com/ascend-api:latest
```

### Production docker-compose

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  api:
    image: ascend-api:latest
    environment:
      ENV: production
      DB_PASSWORD: ${DB_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
    deploy:
      replicas: 3
      restart_policy:
        condition: on-failure
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
ports:
  - "8081:8080"  # Use 8081 instead
```

### Container Won't Start

```bash
# Check logs
docker-compose logs api

# Remove and recreate
docker-compose down
docker-compose up --force-recreate api
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check health
docker-compose exec postgres pg_isready -U ascend

# Restart PostgreSQL
docker-compose restart postgres
```

### Hot-Reload Not Working

```bash
# Rebuild dev image
docker-compose -f docker-compose.yml -f docker-compose.dev.yml build --no-cache api

# Restart with logs
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up api
```

### Out of Disk Space

```bash
# Remove unused images
docker system prune -a

# Remove volumes
docker volume prune

# Check disk usage
docker system df
```

## Performance Optimization

### Resource Limits

Add to docker-compose.yml:

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
```

### Caching

```bash
# Use BuildKit for faster builds
DOCKER_BUILDKIT=1 docker-compose build

# Cache dependencies
docker-compose build --build-arg BUILDKIT_INLINE_CACHE=1
```

## Maintenance

### Backup Database

```bash
# Dump database
docker-compose exec -T postgres pg_dump -U ascend ascend > backup.sql

# Restore database
docker-compose exec -T postgres psql -U ascend ascend < backup.sql
```

### Update Images

```bash
# Pull latest base images
docker-compose pull

# Rebuild with latest
docker-compose build --no-cache

# Restart services
docker-compose up -d
```

### Clean Up

```bash
# Stop and remove everything
./scripts/stop-dev.sh --clean

# Remove dangling images
docker image prune

# Remove all unused resources
docker system prune -a --volumes
```

## Security Notes

### Development
- Default passwords are **NOT** secure
- CORS is wide-open for development
- Debug mode is enabled

### Production
- Change all default passwords
- Restrict CORS origins
- Disable debug mode
- Use secrets management (Docker Secrets, Vault)
- Enable SSL/TLS
- Run as non-root user
- Use security scanning (Trivy, Snyk)

## Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)
- [Redis Docker Hub](https://hub.docker.com/_/redis)
- [MinIO Documentation](https://min.io/docs/)
- [Air - Live Reload](https://github.com/cosmtrek/air)

---

**Happy Coding!** 🚀
