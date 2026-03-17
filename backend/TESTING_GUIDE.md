# Testing Guide - Ascend API

## Overview

This guide covers testing practices, infrastructure, and guidelines for the Ascend API backend.

## Test Structure

Our testing follows the standard Go testing conventions with additional organization:

```
backend/
├── internal/
│   ├── handler/
│   │   ├── auth_handler.go
│   │   └── auth_handler_test.go      # Handler tests
│   ├── service/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go      # Service tests
│   ├── repository/
│   │   ├── user_repository.go
│   │   └── user_repository_test.go   # Repository tests
│   └── middleware/
│       ├── auth.go
│       └── auth_test.go              # Middleware tests
├── pkg/
│   └── auth/
│       ├── password.go
│       └── password_test.go          # Package tests
└── test/
    ├── integration/                  # Integration tests
    ├── fixtures/                     # Test data
    └── helpers/                      # Test utilities
```

## Test Categories

### 1. Unit Tests

Test individual components in isolation with mocked dependencies.

**Location**: `*_test.go` files alongside source files

**Example**:
```go
func TestHashPassword(t *testing.T) {
    tests := []struct {
        name        string
        password    string
        expectError bool
    }{
        {
            name:        "valid password",
            password:    "SecurePassword123!",
            expectError: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            hash, err := auth.HashPassword(tt.password)

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, hash)
            }
        })
    }
}
```

**Run unit tests**:
```bash
make test-unit
# or
go test -short ./...
```

### 2. Integration Tests

Test interactions between multiple components with real dependencies.

**Location**: `test/integration/`

**Naming**: Use `_integration_test.go` suffix

**Example**:
```go
// +build integration

func TestUserAuthenticationFlow_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDatabase(t)
    defer teardownTestDatabase(t, db)

    // Test complete authentication flow
    // ...
}
```

**Run integration tests**:
```bash
make test-integration
# or
go test -run Integration ./...
```

### 3. End-to-End Tests

Test complete user scenarios through the HTTP API.

**Location**: `test/e2e/`

**Run E2E tests**:
```bash
make test-e2e
# or
go test -tags e2e ./test/e2e/...
```

## Test Dependencies

Required testing libraries (automatically installed):

```go
// go.mod
require (
    github.com/stretchr/testify v1.8.4    // Assertions and mocks
    github.com/DATA-DOG/go-sqlmock v1.5.0 // SQL mocking
    github.com/golang/mock v1.6.0          // Mock generation
)
```

Install all dependencies:
```bash
make install-deps
```

## Writing Tests

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"missing domain", "user@", true},
        {"empty email", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Using Mocks

Generate mocks with mockery:

```bash
make mock
```

Use mocks in tests:

```go
type MockAuthService struct {
    mock.Mock
}

func (m *MockAuthService) Login(email, password string) (*dto.AuthResponse, error) {
    args := m.Called(email, password)
    return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
    mockService := new(MockAuthService)
    mockService.On("Login", "test@example.com", "password").
        Return(&dto.AuthResponse{AccessToken: "token"}, nil)

    handler := NewAuthHandler(mockService)
    // ... test handler

    mockService.AssertExpectations(t)
}
```

### Testing HTTP Handlers

Use `httptest` for handler testing:

```go
func TestHealthCheckHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)

    router := gin.New()
    router.GET("/health", healthCheckHandler)

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/health", nil)

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "healthy")
}
```

### Testing Database Operations

Use `sqlmock` for database testing:

```go
func TestUserRepository_Create(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    mock.ExpectExec("INSERT INTO users").
        WithArgs("test@example.com", "hash", "Test User").
        WillReturnResult(sqlmock.NewResult(1, 1))

    repo := NewUserRepository(db)
    err = repo.Create(user)

    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

## Test Coverage

### Generate Coverage Report

```bash
make test-coverage
```

This generates:
- `coverage.out`: Machine-readable coverage data
- `coverage.html`: HTML coverage report

View coverage in browser:
```bash
open coverage.html
```

### Coverage Goals

- **Overall**: 70% minimum
- **Critical paths**: 90% minimum (auth, payments, security)
- **Business logic**: 80% minimum
- **Utilities**: 60% minimum

### Check Coverage

```bash
# Overall coverage
go test -cover ./...

# Detailed coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Coverage for specific package
go test -cover ./internal/service
```

## Benchmarking

Write benchmark tests for performance-critical code:

```go
func BenchmarkHashPassword(b *testing.B) {
    password := "TestPassword123!"
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, _ = HashPassword(password)
    }
}
```

Run benchmarks:

```bash
make benchmark
# or
go test -bench=. -benchmem ./...
```

Example output:
```
BenchmarkHashPassword-8     100   10234567 ns/op   1024 B/op   10 allocs/op
```

## Test Fixtures

Create reusable test data:

```go
// test/fixtures/users.go
package fixtures

func ValidUser() *domain.User {
    return &domain.User{
        Email: "test@example.com",
        Name:  "Test User",
    }
}

func UserWithBodyWeight(weight float64) *domain.User {
    user := ValidUser()
    user.BodyWeight = &weight
    return user
}
```

## Test Helpers

Create helper functions for common test operations:

```go
// test/helpers/db.go
package helpers

func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper()

    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    require.NoError(t, db.AutoMigrate(&domain.User{}, &domain.Session{}))

    return db
}

func TeardownTestDB(t *testing.T, db *gorm.DB) {
    t.Helper()

    sqlDB, err := db.DB()
    require.NoError(t, err)
    require.NoError(t, sqlDB.Close())
}
```

## Continuous Integration

Our CI pipeline runs:

1. **Linting**: `golangci-lint run`
2. **Formatting**: Check `gofmt` compliance
3. **Vetting**: `go vet ./...`
4. **Security**: `gosec` scan
5. **Tests**: All unit and integration tests
6. **Coverage**: Generate and upload coverage report

### GitHub Actions Workflow

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install dependencies
        run: make install-deps

      - name: Lint
        run: make lint

      - name: Test with coverage
        run: make test-coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Best Practices

### 1. Test Naming

- **Test functions**: `TestFunctionName_Scenario`
- **Examples**:
  - `TestHashPassword_ValidInput`
  - `TestUserRepository_Create_DuplicateEmail`
  - `TestAuthHandler_Login_InvalidCredentials`

### 2. AAA Pattern

Structure tests using Arrange-Act-Assert:

```go
func TestLogin(t *testing.T) {
    // Arrange
    user := createTestUser(t)
    service := NewAuthService(mockRepo)

    // Act
    result, err := service.Login(user.Email, "password")

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, user.Email, result.Email)
}
```

### 3. Parallel Tests

Run independent tests in parallel:

```go
func TestConcurrentOperations(t *testing.T) {
    t.Parallel() // Mark as parallelizable

    // Test logic
}
```

### 4. Cleanup with t.Cleanup

Use `t.Cleanup` for guaranteed cleanup:

```go
func TestWithCleanup(t *testing.T) {
    resource := allocateResource()
    t.Cleanup(func() {
        resource.Close()
    })

    // Test logic
}
```

### 5. Subtests

Use subtests for better organization:

```go
func TestAuthFlow(t *testing.T) {
    t.Run("registration", func(t *testing.T) {
        // Registration tests
    })

    t.Run("login", func(t *testing.T) {
        // Login tests
    })
}
```

### 6. Test Data Builders

Use builder pattern for complex test data:

```go
type UserBuilder struct {
    user *domain.User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: &domain.User{
            Email: "default@example.com",
            Name:  "Default User",
        },
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) Build() *domain.User {
    return b.user
}

// Usage
user := NewUserBuilder().
    WithEmail("test@example.com").
    Build()
```

## Troubleshooting

### Tests Failing Locally

1. **Ensure dependencies are installed**:
   ```bash
   make install-deps
   ```

2. **Check environment variables**:
   ```bash
   cp .env.example .env.test
   ```

3. **Clean and rebuild**:
   ```bash
   make clean
   make build
   ```

### Database Tests Failing

1. **Use in-memory SQLite for speed**:
   ```go
   db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
   ```

2. **Isolate tests with transactions**:
   ```go
   tx := db.Begin()
   defer tx.Rollback()
   ```

### Flaky Tests

1. **Avoid time-based assertions**:
   ```go
   // Bad
   time.Sleep(100 * time.Millisecond)

   // Good
   <-done // Wait for channel signal
   ```

2. **Use test timeouts**:
   ```go
   func TestWithTimeout(t *testing.T) {
       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
       defer cancel()

       // Test with context
   }
   ```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Comments](https://go.dev/blog/subtests)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)

## Quick Reference

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run specific test
go test -run TestFunctionName ./path/to/package

# Run tests in verbose mode
go test -v ./...

# Run tests with race detector
go test -race ./...

# Generate mocks
make mock

# Run benchmarks
make benchmark

# Check code quality
make check
```

---

**Last Updated**: Sprint 1 - Testing Infrastructure Setup
