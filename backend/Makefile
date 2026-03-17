# Ascend API - Makefile
# Provides convenient commands for development, testing, and deployment

.PHONY: help
help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: run
run: ## Run the application in development mode
	@echo "Starting Ascend API..."
	go run cmd/api/main.go

.PHONY: dev
dev: ## Run the application with live reload (requires air)
	@echo "Starting development server with live reload..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air for live reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

.PHONY: build
build: ## Build the application binary
	@echo "Building application..."
	go build -o bin/ascend-api cmd/api/main.go
	@echo "Binary built: bin/ascend-api"

.PHONY: install-deps
install-deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Testing
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-unit
test-unit: ## Run only unit tests
	@echo "Running unit tests..."
	go test -v -short ./...

.PHONY: test-integration
test-integration: ## Run only integration tests
	@echo "Running integration tests..."
	go test -v -run Integration ./...

.PHONY: test-watch
test-watch: ## Run tests in watch mode (requires gotestsum)
	@echo "Running tests in watch mode..."
	@if ! command -v gotestsum > /dev/null; then \
		echo "Installing gotestsum..."; \
		go install gotest.tools/gotestsum@latest; \
	fi
	gotestsum --watch ./...

.PHONY: benchmark
benchmark: ## Run benchmark tests
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Code Quality
.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...

.PHONY: format
format: ## Format code with gofmt and goimports
	@echo "Formatting code..."
	gofmt -w -s .
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	@if ! command -v staticcheck > /dev/null; then \
		echo "Installing staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	staticcheck ./...

.PHONY: security
security: ## Run security checks with gosec
	@echo "Running security checks..."
	@if ! command -v gosec > /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	gosec -exclude=G104 ./...

.PHONY: check
check: format vet lint security test ## Run all checks (format, vet, lint, security, test)

# Database
.PHONY: db-create
db-create: ## Create database
	@echo "Creating database..."
	createdb ascend_db || echo "Database may already exist"

.PHONY: db-drop
db-drop: ## Drop database (WARNING: destructive)
	@echo "Dropping database..."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		dropdb ascend_db; \
	fi

.PHONY: db-reset
db-reset: db-drop db-create ## Reset database (drop and create)

# Check if DATABASE_URL is set
ifndef DATABASE_URL
DATABASE_URL = $(shell grep DATABASE_URL .env | cut -d '=' -f2)
endif

.PHONY: migrate-install
migrate-install: ## Install golang-migrate CLI tool
	@echo "Installing golang-migrate..."
	@if ! command -v migrate > /dev/null; then \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		echo "golang-migrate installed successfully"; \
	else \
		echo "golang-migrate already installed"; \
	fi

.PHONY: migrate-up
migrate-up: migrate-install ## Run all pending migrations
	@echo "Running migrations..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL not set. Source .env file or set DATABASE_URL"; \
		exit 1; \
	fi
	migrate -path migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: migrate-install ## Rollback last migration
	@echo "Rolling back last migration..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL not set. Source .env file or set DATABASE_URL"; \
		exit 1; \
	fi
	migrate -path migrations -database "$(DATABASE_URL)" down 1

.PHONY: migrate-down-all
migrate-down-all: migrate-install ## Rollback ALL migrations (WARNING: destructive)
	@echo "WARNING: This will rollback ALL migrations!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path migrations -database "$(DATABASE_URL)" down -all; \
	fi

.PHONY: migrate-force
migrate-force: migrate-install ## Force migration version (usage: make migrate-force version=5)
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=5"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(version)..."
	migrate -path migrations -database "$(DATABASE_URL)" force $(version)

.PHONY: migrate-version
migrate-version: migrate-install ## Show current migration version
	@echo "Current migration version:"
	@migrate -path migrations -database "$(DATABASE_URL)" version || echo "No migrations applied"

.PHONY: migrate-status
migrate-status: migrate-version ## Alias for migrate-version

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create name=add_users_column)
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=add_users_column"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@if ! command -v migrate > /dev/null; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)
	@echo "Migration files created in migrations/"

.PHONY: migrate-validate
migrate-validate: migrate-install ## Validate migration files
	@echo "Validating migrations..."
	@for file in migrations/*.sql; do \
		echo "Checking $$file..."; \
		if ! grep -q ";" $$file; then \
			echo "WARNING: $$file may not have SQL statements"; \
		fi; \
	done
	@echo "Validation complete"

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t ascend-api:latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env ascend-api:latest

.PHONY: docker-compose-up
docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services..."
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

# Utilities
.PHONY: generate
generate: ## Run go generate
	@echo "Running go generate..."
	go generate ./...

.PHONY: mock
mock: ## Generate mocks (requires mockery)
	@echo "Generating mocks..."
	@if ! command -v mockery > /dev/null; then \
		echo "Installing mockery..."; \
		go install github.com/vektra/mockery/v2@latest; \
	fi
	mockery --all --dir internal --output internal/mocks

.PHONY: clean
clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	go clean
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf tmp/

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: deps-audit
deps-audit: ## Audit dependencies for vulnerabilities
	@echo "Auditing dependencies..."
	go list -json -m all | nancy sleuth

.PHONY: env-setup
env-setup: ## Set up environment file from example
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env with your configuration"; \
	else \
		echo ".env already exists"; \
	fi

# Production
.PHONY: build-prod
build-prod: ## Build production binary with optimizations
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-w -s -X main.Version=$$(git describe --tags --always)" \
		-o bin/ascend-api-linux-amd64 \
		cmd/api/main.go
	@echo "Production binary built: bin/ascend-api-linux-amd64"

.PHONY: deploy
deploy: build-prod ## Deploy to production (requires proper configuration)
	@echo "Deploying to production..."
	@echo "This is a placeholder. Implement your deployment strategy here."

# CI/CD
.PHONY: ci
ci: lint vet security test-coverage ## Run CI pipeline locally

.PHONY: pre-commit
pre-commit: format vet lint ## Run pre-commit checks
	@echo "Pre-commit checks passed!"

# Documentation
.PHONY: docs
docs: ## Generate API documentation
	@echo "Generating documentation..."
	@if ! command -v swag > /dev/null; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/api/main.go

.PHONY: docs-serve
docs-serve: docs ## Serve API documentation
	@echo "Serving documentation at http://localhost:8080/swagger/index.html"
	@echo "Start the server with 'make run' in another terminal"

# Health checks
.PHONY: health
health: ## Check application health
	@echo "Checking application health..."
	@curl -s http://localhost:8080/health | jq . || echo "Application not running or jq not installed"

.PHONY: metrics
metrics: ## View application metrics
	@echo "Fetching metrics..."
	@curl -s http://localhost:8080/metrics || echo "Application not running"

# Default target
.DEFAULT_GOAL := help
