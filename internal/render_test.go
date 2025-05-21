package internal

import (
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
