package errors

import (
	"errors"
	"fmt"
)

// Application error types for proper HTTP status code mapping
var (
	// ErrNotFound represents a resource not found error (HTTP 404)
	ErrNotFound = errors.New("resource not found")

	// ErrConflict represents a conflict error, e.g., duplicate resource (HTTP 409)
	ErrConflict = errors.New("resource already exists")

	// ErrUnauthorized represents an authentication error (HTTP 401)
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden represents an authorization error (HTTP 403)
	ErrForbidden = errors.New("forbidden")

	// ErrBadRequest represents a validation or malformed request error (HTTP 400)
	ErrBadRequest = errors.New("bad request")

	// ErrInternal represents an internal server error (HTTP 500)
	ErrInternal = errors.New("internal server error")
)

// ConflictError creates a new conflict error with a custom message
func ConflictError(message string) error {
	return fmt.Errorf("%w: %s", ErrConflict, message)
}

// NotFoundError creates a new not found error with a custom message
func NotFoundError(message string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, message)
}

// UnauthorizedError creates a new unauthorized error with a custom message
func UnauthorizedError(message string) error {
	return fmt.Errorf("%w: %s", ErrUnauthorized, message)
}

// ForbiddenError creates a new forbidden error with a custom message
func ForbiddenError(message string) error {
	return fmt.Errorf("%w: %s", ErrForbidden, message)
}

// BadRequestError creates a new bad request error with a custom message
func BadRequestError(message string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, message)
}

// InternalError creates a new internal error with a custom message
func InternalError(message string) error {
	return fmt.Errorf("%w: %s", ErrInternal, message)
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsBadRequest checks if an error is a bad request error
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}
