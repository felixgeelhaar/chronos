package main

import "fmt"

// ExitCode enumerates the process exit codes chronos uses. They follow
// CLI convention: 0 = success, 2 = usage, 3 = not-found, anything else =
// generic error. Stable codes make scripts that wrap chronos easier to
// write.
type ExitCode int

const (
	ExitSuccess  ExitCode = 0
	ExitError    ExitCode = 1
	ExitUsage    ExitCode = 2
	ExitNotFound ExitCode = 3
)

// ChronosError carries an exit code, a user-visible message, an optional
// underlying cause (revealed when CHRONOS_VERBOSE is set), and an optional
// hint that points at how to fix the problem. Subcommands return
// ChronosErrors to communicate intent to main; main translates them into
// stderr output and an exit code.
type ChronosError struct {
	Code    ExitCode
	Message string
	Cause   error
	Hint    string
}

func (e *ChronosError) Error() string { return e.Message }

func (e *ChronosError) Unwrap() error { return e.Cause }

// NewUserError reports a usage problem (bad flag, missing required value).
func NewUserError(format string, args ...any) *ChronosError {
	return &ChronosError{
		Code:    ExitUsage,
		Message: fmt.Sprintf(format, args...),
		Hint:    "run 'chronos help' for usage",
	}
}

// NewSystemError wraps an unexpected internal failure.
func NewSystemError(cause error, format string, args ...any) *ChronosError {
	return &ChronosError{
		Code:    ExitError,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

// NewNotFoundError reports a configuration or lookup miss (unknown
// adapter, etc.).
func NewNotFoundError(format string, args ...any) *ChronosError {
	return &ChronosError{
		Code:    ExitNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}
