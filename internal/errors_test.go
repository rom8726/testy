package internal

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	// Test creating a new error
	err := NewError(ErrNotFound, "TestOp", "test message")

	// Check error type
	if err.Err != ErrNotFound {
		t.Errorf("Expected Err to be ErrNotFound, got %v", err.Err)
	}

	// Check operation
	if err.Op != "TestOp" {
		t.Errorf("Expected Op to be 'TestOp', got '%s'", err.Op)
	}

	// Check message
	if err.Message != "test message" {
		t.Errorf("Expected Message to be 'test message', got '%s'", err.Message)
	}

	// Check context map is initialized
	if err.Context == nil {
		t.Error("Expected Context to be initialized, got nil")
	}
}

func TestError_Error(t *testing.T) {
	// Test with message
	err1 := NewError(ErrNotFound, "TestOp", "test message")
	if err1.Error() != "test message" {
		t.Errorf("Expected Error() to return 'test message', got '%s'", err1.Error())
	}

	// Test with op but no message
	err2 := NewError(ErrNotFound, "TestOp", "")
	if err2.Error() != "TestOp: not found" {
		t.Errorf("Expected Error() to return 'TestOp: not found', got '%s'", err2.Error())
	}

	// Test with no op and no message
	err3 := NewError(ErrNotFound, "", "")
	if err3.Error() != "not found" {
		t.Errorf("Expected Error() to return 'not found', got '%s'", err3.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	// Create a new error
	err := NewError(ErrNotFound, "TestOp", "test message")

	// Unwrap the error
	unwrapped := err.Unwrap()

	// Check that the unwrapped error is the original error
	if unwrapped != ErrNotFound {
		t.Errorf("Expected Unwrap() to return ErrNotFound, got %v", unwrapped)
	}
}

func TestError_Is(t *testing.T) {
	// Create a new error
	err := NewError(ErrNotFound, "TestOp", "test message")

	// Check that Is returns true for the original error
	if !err.Is(ErrNotFound) {
		t.Error("Expected Is(ErrNotFound) to return true, got false")
	}

	// Check that Is returns false for a different error
	if err.Is(ErrInvalidInput) {
		t.Error("Expected Is(ErrInvalidInput) to return false, got true")
	}

	// Check that errors.Is works with our custom error
	if !errors.Is(err, ErrNotFound) {
		t.Error("Expected errors.Is(err, ErrNotFound) to return true, got false")
	}
}

func TestError_WithContext(t *testing.T) {
	// Create a new error
	err := NewError(ErrNotFound, "TestOp", "test message")

	// Add context
	err = err.WithContext("key1", "value1")
	err = err.WithContext("key2", 42)

	// Check that the context was added
	if val, ok := err.Context["key1"]; !ok || val != "value1" {
		t.Errorf("Expected Context['key1'] to be 'value1', got %v", val)
	}

	if val, ok := err.Context["key2"]; !ok || val != 42 {
		t.Errorf("Expected Context['key2'] to be 42, got %v", val)
	}

	// Check that WithContext returns the same error
	err2 := err.WithContext("key3", true)
	if err2 != err {
		t.Error("Expected WithContext to return the same error instance")
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types are defined
	if ErrNotFound == nil {
		t.Error("ErrNotFound is nil")
	}

	if ErrInvalidInput == nil {
		t.Error("ErrInvalidInput is nil")
	}

	if ErrDatabase == nil {
		t.Error("ErrDatabase is nil")
	}

	if ErrHTTP == nil {
		t.Error("ErrHTTP is nil")
	}

	if ErrMock == nil {
		t.Error("ErrMock is nil")
	}

	if ErrInternal == nil {
		t.Error("ErrInternal is nil")
	}

	// Test that all error types have different messages
	errorMessages := map[string]bool{}
	for _, err := range []error{
		ErrNotFound,
		ErrInvalidInput,
		ErrDatabase,
		ErrHTTP,
		ErrMock,
		ErrInternal,
	} {
		msg := err.Error()
		if errorMessages[msg] {
			t.Errorf("Duplicate error message: %s", msg)
		}
		errorMessages[msg] = true
	}
}
