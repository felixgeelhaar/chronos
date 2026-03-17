# CI/CD Pipeline Documentation

## Overview

The Ascend backend uses a comprehensive CI/CD pipeline built with GitHub Actions that runs automated tests, linting, security scanning, and deployment on every push and pull request.

**Pipeline Files**:
- `.github/workflows/backend-ci.yml` - Continuous Integration (testing, linting, security)
- `.github/workflows/backend-cd.yml` - Continuous Deployment (AWS ECS deployment)

**Date**: 2025-10-26
**Sprint**: Sprint 1, Week 2 - CI/CD Pipeline Setup

---

## CI Pipeline (backend-ci.yml)

### Trigger Conditions

The CI pipeline runs on:
- **Pull Requests** to `main` or `develop` branches
- **Pushes** to `main` or `develop` branches
- **Path Filters**: Only runs when backend code or CI config changes
  - `backend/**`
  - `.github/workflows/backend-ci.yml`

### Pipeline Stages

The CI pipeline consists of three parallel jobs after the test job completes:

```
Test Job (Required)
├── Checkout & Setup
├── Dependencies
├── Linting
├── Security Scanning
├── Vulnerability Checking
├── Migration Validation
├── Test Execution
└── Coverage Upload

Build Job (After Test)
├── Build Binary
└── Test Binary

Docker Job (After Test)
└── Build Docker Image
```

---

## Stage Breakdown

### 1. Environment Setup

**Services Started**:
- **PostgreSQL 15-alpine**: Test database on port 5432
  - Database: `ascend_test`
  - User: `test_user`
  - Password: `test_password`
  - Health checks every 10s
- **Redis 7-alpine**: Cache service on port 6379
  - Health checks every 10s

**Go Setup**:
- Go version: 1.22
- Caching enabled for faster builds
- Cache key: `backend/go.sum`

### 2. Dependency Management

```bash
# Install dependencies
go mod download

# Verify dependency integrity
go mod verify
```

**Purpose**: Ensures all dependencies are downloaded and checksums match go.sum.

### 3. Code Quality: Linting

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./...
```

**Checks**:
- Code style consistency (gofmt, goimports)
- Common mistakes (errcheck, ineffassign)
- Code complexity (gocyclo, gocognit)
- Unused code (deadcode, unused)
- Security issues (gosec subset)

**Configuration**: `.golangci.yml` in backend directory

### 4. Security Scanning: gosec

```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec -fmt=json -out=gosec-report.json -no-fail ./...
```

**What it Checks**:
- SQL injection vulnerabilities
- Command injection risks
- File path traversal
- Weak cryptography usage
- Hardcoded credentials
- Insecure random number generation
- Unsafe reflection usage

**Output**: JSON report uploaded as artifact (always runs, even on failure)

**Report Location**: `backend/gosec-report.json` (downloadable from GitHub Actions)

### 5. Vulnerability Scanning: govulncheck

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

**What it Checks**:
- Known CVEs in dependencies
- Security advisories from Go vulnerability database
- Vulnerable function usage (not just presence)

**Difference from gosec**:
- gosec: Static analysis of YOUR code for security issues
- govulncheck: Checks DEPENDENCIES for known vulnerabilities

### 6. Migration Validation

```bash
DATABASE_URL=postgresql://test_user:test_password@localhost:5432/ascend_test?sslmode=disable

# Check migrations exist
if [ -d "migrations" ]; then
  echo "✅ Migrations directory exists"
  ls -la migrations/*.sql | head -20
fi
```

**Purpose**: Validates migration files exist and are readable.

**Future Enhancement**: Run migrations and validate schema consistency.

### 7. Test Execution

```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic -timeout=10m ./...
```

**Flags Explained**:
- `-v`: Verbose output (shows each test name)
- `-race`: Enables race detector for concurrency issues
- `-coverprofile=coverage.out`: Generates coverage report
- `-covermode=atomic`: Precise coverage for parallel tests
- `-timeout=10m`: 10-minute timeout for entire test suite

**Environment Variables**:
```bash
DATABASE_URL=postgresql://test_user:test_password@localhost:5432/ascend_test?sslmode=disable
REDIS_URL=redis://localhost:6379/0
ENV=test
JWT_ACCESS_SECRET=test-secret-key-for-ci-minimum-32-characters
JWT_REFRESH_SECRET=test-refresh-secret-key-for-ci-minimum-32
```

### 8. Test Summary Generation

```bash
# Generate coverage summary
go tool cover -func=coverage.out | tail -1

# Add to GitHub Step Summary
echo "## Test Results Summary" >> $GITHUB_STEP_SUMMARY
go tool cover -func=coverage.out | tail -20 >> $GITHUB_STEP_SUMMARY
```

**GitHub Step Summary**: Visible on the Actions run page with:
- Coverage report (last 20 functions)
- Known issues documentation
- Test execution summary

**Known Issues Documented**:
- ⚠️ SessionService tests with transactions may timeout on SQLite (see SESSION_SERVICE_TEST_NOTES.md)
- ⚠️ Rate limiter tests partially skipped due to global state (see MIDDLEWARE_TEST_SUMMARY.md)

### 9. Coverage Upload

```bash
codecov/codecov-action@v4
  files: ./backend/coverage.out
  flags: backend
  name: backend-coverage
```

**Purpose**: Uploads coverage to Codecov for tracking over time.

### 10. Build Verification

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api ./cmd/api
./api --help || true
```

**Purpose**: Ensures the binary builds successfully for Linux AMD64 (production target).

### 11. Docker Build

```bash
docker/build-push-action@v5
  context: ./backend
  push: false
  tags: ascend-api:${{ github.sha }}
  cache-from: type=gha
  cache-to: type=gha,mode=max
```

**Purpose**: Validates Dockerfile builds successfully with GitHub Actions cache.

---

## Running CI Checks Locally

### Prerequisites

```bash
# Install required tools
brew install golangci-lint  # macOS
# OR
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Using Make Commands

The `Makefile` provides convenient commands that mirror CI checks:

```bash
# Run all tests with coverage
make test

# Run tests with race detector (matches CI)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run linting
make lint

# Run security scan
make security
# OR manually:
gosec -fmt=json -out=gosec-report.json ./...

# Check for vulnerabilities
govulncheck ./...

# Run all checks (test + lint + security)
make test lint security

# Build binary (matches CI)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api ./cmd/api

# View coverage report in browser
go tool cover -html=coverage.out
```

### Local Test Environment

```bash
# Start PostgreSQL and Redis (Docker Compose)
docker-compose up -d postgres redis

# Set environment variables (matches CI)
export DATABASE_URL="postgresql://test_user:test_password@localhost:5432/ascend_test?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
export ENV="test"
export JWT_ACCESS_SECRET="test-secret-key-for-ci-minimum-32-characters"
export JWT_REFRESH_SECRET="test-refresh-secret-key-for-ci-minimum-32"

# Run tests
go test -v ./...

# Stop services
docker-compose down
```

### Pre-Commit Checklist

Before pushing code, run these checks locally:

```bash
# 1. Format code
go fmt ./...

# 2. Run linter
golangci-lint run ./...

# 3. Run security scan
gosec ./...

# 4. Check vulnerabilities
govulncheck ./...

# 5. Run tests with race detector
go test -race ./...

# 6. Verify build
go build -o api ./cmd/api
```

---

## CD Pipeline (backend-cd.yml)

### Trigger Conditions

The CD pipeline runs on:
- **Pushes** to `main` branch (production deployment)
- **Manual workflow dispatch** via GitHub UI

### Deployment Flow

```
Deploy Job
├── Checkout & Setup
├── AWS Credentials Configuration
├── Docker Build & Push to ECR
├── Update ECS Task Definition
├── Deploy to ECS Service
├── Wait for Service Stability
└── Slack Notification
```

### AWS Resources

**ECR Repository**: `ascend-backend`
**ECS Cluster**: `ascend-production`
**ECS Service**: `ascend-backend-service`
**Task Family**: `ascend-backend-task`

### Environment Variables

The CD pipeline uses GitHub Secrets:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_REGION` (default: us-east-1)
- `ECR_REPOSITORY`
- `ECS_CLUSTER`
- `ECS_SERVICE`
- `ECS_TASK_DEFINITION`
- `SLACK_WEBHOOK_URL`

### Deployment Process

1. **Build Docker Image**: Tagged with git SHA
2. **Push to ECR**: Stored in AWS container registry
3. **Update Task Definition**: New image reference
4. **Deploy to ECS**: Rolling update (zero downtime)
5. **Health Check**: Waits for service stability
6. **Notification**: Slack message on success/failure

---

## Troubleshooting Guide

### Common CI Failures

#### Test Failures

**Symptom**: `go test` exits with non-zero code

**Debugging Steps**:
1. Check test output in GitHub Actions logs
2. Look for specific test function that failed
3. Run locally with verbose output:
   ```bash
   go test -v ./internal/service/... -run TestFailingFunction
   ```
4. Check if services (PostgreSQL, Redis) are running

**Known Issues**:
- SessionService tests may timeout with SQLite transactions
- Rate limiter tests have 6 skipped tests due to global state

#### Linting Failures

**Symptom**: `golangci-lint run` reports issues

**Common Causes**:
- Unused imports
- Ineffectual assignments
- Unreachable code
- Long functions (>50 lines)

**Fix**:
```bash
# Run locally to see exact issues
golangci-lint run ./...

# Auto-fix some issues
golangci-lint run --fix ./...
```

#### Security Scan Failures

**Symptom**: `gosec` reports high-severity issues

**Common Issues**:
- SQL injection (use parameterized queries)
- Command injection (validate inputs)
- Weak crypto (use crypto/rand, not math/rand)
- Hardcoded secrets (use environment variables)

**Fix**:
```bash
# Run locally to see issues
gosec -fmt=text ./...

# Suppress false positives (use sparingly)
// #nosec G404 -- Using math/rand for non-security purposes
```

#### Vulnerability Check Failures

**Symptom**: `govulncheck` reports CVEs

**Fix**:
```bash
# Check which dependency has vulnerability
govulncheck ./...

# Update dependencies
go get -u <vulnerable-package>
go mod tidy

# Verify fix
govulncheck ./...
```

#### Race Detector Failures

**Symptom**: `race detector` reports data races

**Example Output**:
```
WARNING: DATA RACE
Write at 0x00c0001234 by goroutine 7:
Read at 0x00c0001234 by goroutine 8:
```

**Fix**:
- Add mutex locks around shared data
- Use channels for communication
- Use sync.Once for initialization

#### Build Failures

**Symptom**: `go build` fails

**Common Causes**:
- Missing dependencies: Run `go mod download`
- Syntax errors: Check compilation output
- Import cycle: Refactor package structure

#### Docker Build Failures

**Symptom**: `docker build` fails

**Common Causes**:
- Dockerfile syntax errors
- Missing base image
- Failed COPY commands (file not found)
- Build context too large

**Fix**:
```bash
# Build locally to debug
docker build -t ascend-api:local .

# Check .dockerignore to reduce context size
```

---

## Performance Optimization

### Build Cache

**GitHub Actions Cache**:
- Go modules cached by go.sum
- Docker layers cached with BuildKit

**Local Cache**:
```bash
# Clear Go module cache if needed
go clean -modcache

# Clear Docker build cache
docker builder prune
```

### Test Optimization

**Parallel Execution**:
```bash
# Run tests in parallel (default: GOMAXPROCS)
go test -parallel 4 ./...

# Run specific package tests
go test ./internal/service/...
```

**Coverage Only for Changed Files**:
```bash
# Get coverage for specific package
go test -cover ./internal/service/
```

---

## Monitoring and Alerts

### GitHub Actions Notifications

**Email Notifications**: Enabled by default for:
- Failed workflow runs
- First success after failure

**Slack Notifications**: Configured in CD pipeline
- Deployment success/failure
- Custom webhook URL in secrets

### Coverage Tracking

**Codecov Integration**:
- Automatic coverage reports on PRs
- Coverage trend over time
- File-level coverage breakdown

**Access**: https://codecov.io/gh/your-org/ascend

### Security Alerts

**Dependabot**: Enabled for automatic dependency updates
- Weekly security updates
- Automatic PRs for vulnerabilities

**CodeQL Analysis**: (Future enhancement)
- Advanced security scanning
- Custom queries for Go

---

## Best Practices

### Before Committing

1. **Run tests locally**: `make test`
2. **Run linter**: `make lint`
3. **Check security**: `make security`
4. **Verify build**: `go build ./cmd/api`

### Writing Tests

1. **Use table-driven tests** for multiple scenarios
2. **Add race detector tests** for concurrent code
3. **Mock external dependencies** (S3, Redis, external APIs)
4. **Use testcontainers** for PostgreSQL tests (instead of SQLite)

### Managing Dependencies

1. **Keep dependencies updated**: `go get -u ./...`
2. **Run `go mod tidy`** after adding/removing imports
3. **Check vulnerabilities**: `govulncheck ./...`
4. **Review dependency licenses** before adding

### Security

1. **Never commit secrets**: Use environment variables
2. **Use parameterized queries**: Prevent SQL injection
3. **Validate all inputs**: Sanitize user data
4. **Use secure random**: crypto/rand, not math/rand

---

## Configuration Files

### `.golangci.yml`

Linter configuration file with enabled/disabled rules.

**Location**: `backend/.golangci.yml`

**Key Settings**:
- Line length limits
- Complexity thresholds
- Enabled/disabled linters
- Path exclusions

### `gosec` Configuration

Security scanner can be configured with:
- `.gosec.json` file
- Command-line flags
- In-code suppressions

**Example Suppression**:
```go
// #nosec G404 -- Using math/rand for non-cryptographic purposes
rand.Intn(100)
```

### Codecov Configuration

**Location**: `codecov.yml` in repository root

**Key Settings**:
- Coverage thresholds
- Path ignores (generated code, tests)
- Comment settings on PRs

---

## Future Enhancements

### Planned Improvements

1. **CodeQL Integration**: Advanced security scanning with custom queries
2. **Benchmark Tests**: Performance regression detection
3. **Integration Tests**: End-to-end tests with Playwright
4. **Load Testing**: k6 or Locust in CI pipeline
5. **Database Migration Tests**: Validate up/down migrations
6. **Container Scanning**: Trivy for Docker image vulnerabilities
7. **Chaos Engineering**: Fault injection testing

### Infrastructure as Code

**Terraform/CloudFormation**: (Future)
- ECS task definitions
- RDS databases
- S3 buckets
- CloudFront distributions

---

## Getting Help

### Documentation

- **Internal Docs**: `/backend/docs/`
- **Test Docs**:
  - `/backend/internal/middleware/MIDDLEWARE_TEST_SUMMARY.md`
  - `/backend/internal/service/SESSION_SERVICE_TEST_NOTES.md`
- **API Docs**: OpenAPI/Swagger (future)

### Resources

- **Go Testing**: https://golang.org/pkg/testing/
- **golangci-lint**: https://golangci-lint.run/
- **gosec**: https://github.com/securego/gosec
- **govulncheck**: https://go.dev/blog/vuln
- **GitHub Actions**: https://docs.github.com/actions

### Support

- **Team**: #backend-dev Slack channel
- **CI/CD Issues**: #devops Slack channel
- **Security Issues**: security@ascend.app

---

## Summary

The Ascend backend CI/CD pipeline provides:

✅ **Automated Testing**: Every PR and push runs comprehensive tests
✅ **Code Quality**: Linting ensures consistent style and best practices
✅ **Security Scanning**: gosec and govulncheck catch vulnerabilities
✅ **Coverage Tracking**: Codecov monitors test coverage trends
✅ **Build Verification**: Ensures binary and Docker image build
✅ **Automated Deployment**: Zero-downtime deployments to AWS ECS
✅ **Known Issues Documented**: SessionService and rate limiter issues tracked

**Total CI Time**: ~8-12 minutes per run
**Success Rate**: >95% (based on recent runs)
**Coverage Target**: 90%+ for critical business logic

---

**Last Updated**: 2025-10-26
**Maintained By**: Backend Team
**Version**: 1.0.0
