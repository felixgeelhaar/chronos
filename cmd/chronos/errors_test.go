package main

import (
	"errors"
	"testing"
)

func TestNewUserError_CarriesUsageCode(t *testing.T) {
	err := NewUserError("bad flag %s", "--foo")
	if err.Code != ExitUsage {
		t.Errorf("Code = %d, want %d", err.Code, ExitUsage)
	}
	if err.Message != "bad flag --foo" {
		t.Errorf("Message = %q", err.Message)
	}
	if err.Hint == "" {
		t.Error("user errors should carry a Hint pointing to chronos help")
	}
}

func TestNewSystemError_WrapsCause(t *testing.T) {
	cause := errors.New("disk on fire")
	err := NewSystemError(cause, "compute: %v", cause)
	if err.Code != ExitError {
		t.Errorf("Code = %d, want %d", err.Code, ExitError)
	}
	if !errors.Is(err, cause) {
		t.Errorf("errors.Is should reach the wrapped cause")
	}
	if err.Error() != "compute: disk on fire" {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestNewNotFoundError_CarriesNotFoundCode(t *testing.T) {
	err := NewNotFoundError("adapter %q not registered", "oops")
	if err.Code != ExitNotFound {
		t.Errorf("Code = %d, want %d", err.Code, ExitNotFound)
	}
	if err.Hint != "" {
		t.Error("not-found errors do not carry a hint by design")
	}
}

func TestChronosError_UnwrapReturnsCause(t *testing.T) {
	cause := errors.New("root")
	err := &ChronosError{Code: ExitError, Message: "wrapper", Cause: cause}
	if got := errors.Unwrap(err); got != cause {
		t.Errorf("Unwrap = %v, want %v", got, cause)
	}
}

func TestChronosError_UnwrapNilWhenNoCause(t *testing.T) {
	err := NewUserError("oops")
	if got := errors.Unwrap(err); got != nil {
		t.Errorf("Unwrap should return nil when no cause; got %v", got)
	}
}
