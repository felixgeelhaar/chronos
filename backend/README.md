# Ascend API - Backend Service

Go-based REST API for the Ascend weightlifting performance tracking platform.

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Gin 1.10+
- **ORM:** GORM 1.25+
- **Database:** PostgreSQL 15+
- **Cache:** Redis 7+
- **Logging:** zerolog

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── middleware/              # HTTP middleware (CORS, logging, auth)
│   ├── domain/                  # Domain models (entities)
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic
│   ├── handler/                 # HTTP handlers (controllers)
│   ├── dto/                     # Data transfer objects
│   ├── util/                    # Utility functions
│   └── worker/                  # Background workers
├── pkg/                         # Shared packages
│   ├── logger/                  # Logging utilities
│   ├── errors/                  # Custom error types
│   └── response/                # Standard API responses
├── migrations/                  # Database migrations
└── tests/                       # Test files
    ├── unit/
    ├── integration/
    └── e2e/
```

## Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 15+
- Redis 7+ (optional for development)
- Docker & Docker Compose (for containerized setup)

### Installation

1. **Clone the repository:**
   ```bash
   cd backend
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run with Docker Compose (recommended):**
   ```bash
   docker-compose up --build
   ```

5. **Or run locally:**
   ```bash
   # Ensure PostgreSQL and Redis are running
   go run cmd/api/main.go
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/service/...
```

### Building

```bash
# Build for current platform
go build -o api ./cmd/api

# Build for Linux (for deployment)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api ./cmd/api
```

## API Endpoints

### Health Check
```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "ascend-api",
  "version": "1.0.0",
  "time": "2025-10-06T12:00:00Z"
}
```

### API v1
Base URL: `/v1`

```bash
GET /v1/ping
```

## Development

### Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing
- Use `golangci-lint` for linting
- Write tests for all business logic

### Database Migrations

Migrations are located in `migrations/` directory and run automatically on startup.

To create a new migration:
```bash
# Create migration file
touch migrations/XXX_description.sql
```

### Environment Variables

See `.env.example` for all available configuration options.

Key variables:
- `PORT`: Server port (default: 8080)
- `ENV`: Environment (development, staging, production)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `JWT_ACCESS_SECRET`: JWT signing secret (must be changed in production)

## Docker

### Build Image
```bash
docker build -t ascend-api .
```

### Run Container
```bash
docker run -p 8080:8080 \
  -e DATABASE_URL=postgresql://... \
  -e REDIS_URL=redis://... \
  ascend-api
```

### Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down

# Rebuild
docker-compose up --build
```

## Deployment

The API is deployed to AWS ECS Fargate in the `eu-west-1` region for GDPR compliance.

### CI/CD Pipeline

GitHub Actions automatically:
1. Runs tests on every PR
2. Builds Docker image
3. Pushes to AWS ECR
4. Deploys to ECS on merge to `main`

## Monitoring

- **Logging:** Structured JSON logs via zerolog
- **Error Tracking:** Sentry (configured via `SENTRY_DSN`)
- **Metrics:** CloudWatch (in production)
- **Health Checks:** `/health` endpoint monitored by ALB

## Contributing

1. Create a feature branch from `develop`
2. Write tests for new features
3. Ensure all tests pass
4. Submit PR with clear description
5. Request review from backend team

## License

Proprietary - All rights reserved

## Support

For issues or questions, contact the backend team or create an issue in the repository.
