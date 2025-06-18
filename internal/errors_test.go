package internal

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	// Test creating a new error
	err := NewError(ErrNotFound, "TestOp", "test message")

	// Check error type
	if !errors.Is(err.Err, ErrNotFound) {
		t.Errorf("Expected Err to be ErrNotFound, got %v", err.Err)
	}

	// Check operation
	if err.Op != "TestOp" {
		t.Errorf("Expected Op to be 'TestOp', got '%s'", err.Op)
	}

	// Check a message
	if err.Message != "test message" {
		t.Errorf("Expected Message to be 'test message', got '%s'", err.Message)
	}

	// Check the context map is initialized
	if err.Context == nil {
		t.Error("Expected Context to be initialized, got nil")
	}
}

func TestError_ErrorMethod(t *testing.T) {
	root := errors.New("root cause")

	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "only message",
			err:      &Error{Message: "boom"},
			expected: "boom",
		},
		{
			name:     "only op + underlying",
			err:      &Error{Op: "DoWork", Err: root},
			expected: "DoWork: root cause",
		},
		{
			name:     "only underlying",
			err:      &Error{Err: root},
			expected: "root cause",
		},
		{
			name:     "message + op",
			err:      &Error{Message: "boom", Op: "DoWork"},
			expected: "boom: DoWork",
		},
		{
			name:     "message + underlying",
			err:      &Error{Message: "boom", Err: root},
			expected: "boom: root cause",
		},
		{
			name:     "op + underlying",
			err:      &Error{Op: "DoWork", Err: root},
			expected: "DoWork: root cause",
		},
		{
			name:     "message + op + underlying",
			err:      &Error{Message: "boom", Op: "DoWork", Err: root},
			expected: "boom: DoWork: root cause",
		},
		{
			name: "with context (sorted)",
			err: &Error{
				Message: "boom",
				Context: map[string]any{
					"b": 2,
					"a": 1,
				},
			},
			expected: "boom {a=1, b=2}",
		},
		{
			name:     "message already contains root -> no duplication",
			err:      &Error{Message: "root cause happened", Err: root},
			expected: "root cause happened",
		},
		{
			name:     "all fields empty",
			err:      &Error{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
