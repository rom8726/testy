package internal

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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
	var b strings.Builder

	if e.Message != "" {
		b.WriteString(e.Message)
	}

	if e.Op != "" {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(e.Op)
	}

	if e.Err != nil {
		underlying := e.Err.Error()
		if underlying != "" && !strings.Contains(b.String(), underlying) {
			if b.Len() > 0 {
				b.WriteString(": ")
			}
			b.WriteString(underlying)
		}
	}

	if len(e.Context) != 0 {
		keys := make([]string, 0, len(e.Context))
		keysExclm := make([]string, 0, len(e.Context))
		for k := range e.Context {
			if strings.HasPrefix(k, "!") {
				keysExclm = append(keysExclm, k)
			} else {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		sort.Strings(keysExclm)

		keysFinal := append(keys, keysExclm...)

		b.WriteString(" {")
		for i, k := range keysFinal {
			if i > 0 {
				b.WriteString(", ")
			}

			_, _ = fmt.Fprintf(&b, "%s=%v", k, e.Context[k])
		}
		b.WriteString("}")
	}

	return b.String()
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

//func (e *Error) Format(s fmt.State, verb rune) {
//	switch verb {
//	case 'v':
//		if s.Flag('+') {
//			_, _ = io.WriteString(s, e.Error())
//
//			if len(e.Context) > 0 {
//				keys := make([]string, 0, len(e.Context))
//				for k := range e.Context {
//					keys = append(keys, k)
//				}
//				sort.Strings(keys)
//
//				_, _ = io.WriteString(s, " {")
//				for i, k := range keys {
//					if i > 0 {
//						_, _ = io.WriteString(s, ", ")
//					}
//					_, _ = fmt.Fprintf(s, "%s=%v", k, e.Context[k])
//				}
//				_, _ = io.WriteString(s, "}")
//			}
//			return
//		}
//
//		fallthrough
//	case 's':
//		_, _ = io.WriteString(s, e.Error())
//	case 'q':
//		_, _ = fmt.Fprintf(s, "%q", e.Error())
//	}
//}
