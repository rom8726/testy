package internal

import (
	"errors"
	"fmt"
)

// Common error types
var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabase is returned when a database operation fails
	ErrDatabase = errors.New("database error")

	// ErrHTTP is returned when an HTTP operation fails
	ErrHTTP = errors.New("HTTP error")

	// ErrMock is returned when a mock operation fails
	ErrMock = errors.New("mock error")

	// ErrInternal is returned when an internal error occurs
	ErrInternal = errors.New("internal error")
)

// Error represents a standardized error with context
type Error struct {
	// Err is the underlying error
	Err error

	// Op is the operation that caused the error
	Op string

	// Message is a human-readable message
	Message string

	// Context contains additional context about the error
	Context map[string]interface{}
}

// Error returns a string representation of the error
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Op != "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}

	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether the error is of the given type
func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewError creates a new Error
func NewError(err error, op string, message string) *Error {
	return &Error{
		Err:     err,
		Op:      op,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	e.Context[key] = value
	return e
}
