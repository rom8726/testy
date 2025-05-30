package internal

import (
	"reflect"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		ctx      map[string]any
		expected string
	}{
		{
			name:     "simple placeholder replacement",
			input:    "Hello, {{name}}!",
			ctx:      map[string]any{"name": "John"},
			expected: "Hello, John!",
		},
		{
			name:     "missing placeholder in context",
			input:    "Hello, {{name}}!",
			ctx:      map[string]any{},
			expected: "Hello, {{name}}!",
		},
		{
			name:     "multiple placeholders",
			input:    "{{greeting}}, {{name}}! Welcome to {{place}}.",
			ctx:      map[string]any{"greeting": "Hi", "name": "Alice", "place": "Wonderland"},
			expected: "Hi, Alice! Welcome to Wonderland.",
		},
		{
			name:     "partial placeholders in context",
			input:    "{{greeting}}, {{name}}! Welcome to {{place}}.",
			ctx:      map[string]any{"greeting": "Hello"},
			expected: "Hello, {{name}}! Welcome to {{place}}.",
		},
		{
			name:     "no placeholders",
			input:    "Hello, World!",
			ctx:      map[string]any{"name": "Someone"},
			expected: "Hello, World!",
		},
		{
			name:     "invalid placeholder format",
			input:    "Hello {{name",
			ctx:      map[string]any{"name": "World"},
			expected: "Hello {{name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderTemplate(tt.input, tt.ctx)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRenderAny(t *testing.T) {
	ctx := map[string]any{
		"name":    "John",
		"age":     30,
		"city":    "New York",
		"enabled": true,
	}

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "string value",
			input:    "Hello, {{name}}!",
			expected: "Hello, John!",
		},
		{
			name: "map value",
			input: map[string]any{
				"greeting": "Hello, {{name}}!",
				"info":     "{{name}} is {{age}} years old",
			},
			expected: map[string]any{
				"greeting": "Hello, John!",
				"info":     "John is 30 years old",
			},
		},
		{
			name:     "array value",
			input:    []any{"{{name}}", "{{age}}", "{{city}}"},
			expected: []any{"John", "30", "New York"},
		},
		{
			name: "nested structures",
			input: map[string]any{
				"user": map[string]any{
					"name": "{{name}}",
					"age":  "{{age}}",
				},
				"addresses": []any{
					map[string]any{
						"city": "{{city}}",
					},
				},
			},
			expected: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  "30",
				},
				"addresses": []any{
					map[string]any{
						"city": "New York",
					},
				},
			},
		},
		{
			name:     "non-string value",
			input:    42,
			expected: 42,
		},
		{
			name:     "boolean value",
			input:    true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderAny(tt.input, ctx)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestRenderRequest(t *testing.T) {
	ctx := map[string]any{
		"userId":    123,
		"userName":  "john_doe",
		"authToken": "abc123xyz",
		"apiPath":   "/api/v1",
	}

	tests := []struct {
		name     string
		input    RequestSpec
		expected RequestSpec
	}{
		{
			name: "simple path replacement",
			input: RequestSpec{
				Method: "GET",
				Path:   "/users/{{userId}}",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: nil,
			},
			expected: RequestSpec{
				Method: "GET",
				Path:   "/users/123",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: nil,
			},
		},
		{
			name: "headers replacement",
			input: RequestSpec{
				Method: "GET",
				Path:   "/users",
				Headers: map[string]string{
					"Authorization": "Bearer {{authToken}}",
					"X-User-ID":     "{{userId}}",
				},
				Body: nil,
			},
			expected: RequestSpec{
				Method: "GET",
				Path:   "/users",
				Headers: map[string]string{
					"Authorization": "Bearer abc123xyz",
					"X-User-ID":     "123",
				},
				Body: nil,
			},
		},
		{
			name: "body replacement",
			input: RequestSpec{
				Method: "POST",
				Path:   "/users",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]any{
					"name":     "{{userName}}",
					"id":       "{{userId}}",
					"metadata": map[string]any{"token": "{{authToken}}"},
				},
			},
			expected: RequestSpec{
				Method: "POST",
				Path:   "/users",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]any{
					"name":     "john_doe",
					"id":       "123",
					"metadata": map[string]any{"token": "abc123xyz"},
				},
			},
		},
		{
			name: "multiple replacements",
			input: RequestSpec{
				Method: "PUT",
				Path:   "{{apiPath}}/users/{{userId}}",
				Headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer {{authToken}}",
				},
				Body: map[string]any{
					"name": "{{userName}}",
				},
			},
			expected: RequestSpec{
				Method: "PUT",
				Path:   "/api/v1/users/123",
				Headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer abc123xyz",
				},
				Body: map[string]any{
					"name": "john_doe",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderRequest(tt.input, ctx)

			// Check path
			if got.Path != tt.expected.Path {
				t.Errorf("Path: expected %q, got %q", tt.expected.Path, got.Path)
			}

			// Check method
			if got.Method != tt.expected.Method {
				t.Errorf("Method: expected %q, got %q", tt.expected.Method, got.Method)
			}

			// Check headers
			if !reflect.DeepEqual(got.Headers, tt.expected.Headers) {
				t.Errorf("Headers: expected %v, got %v", tt.expected.Headers, got.Headers)
			}

			// Check body
			if !reflect.DeepEqual(got.Body, tt.expected.Body) {
				t.Errorf("Body: expected %v, got %v", tt.expected.Body, got.Body)
			}
		})
	}
}

func TestRenderDBCheck(t *testing.T) {
	ctx := map[string]any{
		"userId":   123,
		"userName": "john_doe",
		"orderId":  "ORD-456",
		"status":   "completed",
	}

	tests := []struct {
		name     string
		input    DBCheck
		expected DBCheck
	}{
		{
			name: "simple query replacement",
			input: DBCheck{
				Query:  "SELECT * FROM users WHERE id = {{userId}}",
				Result: "expected result",
			},
			expected: DBCheck{
				Query:  "SELECT * FROM users WHERE id = 123",
				Result: "expected result",
			},
		},
		{
			name: "multiple replacements in query",
			input: DBCheck{
				Query:  "SELECT * FROM orders WHERE user_id = {{userId}} AND status = '{{status}}'",
				Result: "some expected result",
			},
			expected: DBCheck{
				Query:  "SELECT * FROM orders WHERE user_id = 123 AND status = 'completed'",
				Result: "some expected result",
			},
		},
		{
			name: "complex query with joins",
			input: DBCheck{
				Query:  "SELECT o.id, o.status FROM orders o JOIN users u ON o.user_id = u.id WHERE u.name = '{{userName}}' AND o.id = '{{orderId}}'",
				Result: "complex result",
			},
			expected: DBCheck{
				Query:  "SELECT o.id, o.status FROM orders o JOIN users u ON o.user_id = u.id WHERE u.name = 'john_doe' AND o.id = 'ORD-456'",
				Result: "complex result",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderDBCheck(tt.input, ctx)

			// Check query
			if got.Query != tt.expected.Query {
				t.Errorf("Query: expected %q, got %q", tt.expected.Query, got.Query)
			}

			// Result should remain unchanged
			if !reflect.DeepEqual(got.Result, tt.expected.Result) {
				t.Errorf("Result: expected %v, got %v", tt.expected.Result, got.Result)
			}
		})
	}
}
