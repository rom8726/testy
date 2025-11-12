package internal

import (
	"reflect"
	"testing"
)

func TestRenderTemplateWithOptions_PermissiveMode(t *testing.T) {
	ctx := map[string]any{
		"name": "John",
		"age":  30,
	}

	tests := []struct {
		name     string
		input    string
		opts     RenderOptions
		expected string
		wantErr  bool
	}{
		{
			name:  "existing placeholder",
			input: "Hello, {{name}}!",
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: "Hello, John!",
			wantErr:  false,
		},
		{
			name:  "missing placeholder - keep unchanged",
			input: "Hello, {{missing}}!",
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: "Hello, {{missing}}!",
			wantErr:  false,
		},
		{
			name:  "missing placeholder - use default",
			input: "Hello, {{missing}}!",
			opts: RenderOptions{
				Mode:         RenderModePermissive,
				DefaultValue: "UNKNOWN",
			},
			expected: "Hello, UNKNOWN!",
			wantErr:  false,
		},
		{
			name:  "multiple placeholders mixed",
			input: "{{name}} is {{age}} years old and lives in {{city}}",
			opts: RenderOptions{
				Mode:         RenderModePermissive,
				DefaultValue: "N/A",
			},
			expected: "John is 30 years old and lives in N/A",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplateWithOptions(tt.input, ctx, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplateWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.expected {
				t.Errorf("RenderTemplateWithOptions() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRenderTemplateWithOptions_StrictMode(t *testing.T) {
	ctx := map[string]any{
		"name": "John",
		"age":  30,
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "existing placeholder",
			input:   "Hello, {{name}}!",
			wantErr: false,
		},
		{
			name:    "missing placeholder",
			input:   "Hello, {{missing}}!",
			wantErr: true,
		},
		{
			name:    "multiple existing placeholders",
			input:   "{{name}} is {{age}} years old",
			wantErr: false,
		},
		{
			name:    "one missing in multiple placeholders",
			input:   "{{name}} lives in {{city}}",
			wantErr: true,
		},
	}

	opts := RenderOptions{
		Mode: RenderModeStrict,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RenderTemplateWithOptions(tt.input, ctx, opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplateWithOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRenderAnyWithOptions(t *testing.T) {
	ctx := map[string]any{
		"name":    "John",
		"age":     30,
		"city":    "New York",
		"enabled": true,
	}

	tests := []struct {
		name     string
		input    any
		opts     RenderOptions
		expected any
		wantErr  bool
	}{
		{
			name:  "string value - permissive",
			input: "Hello, {{name}}!",
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: "Hello, John!",
			wantErr:  false,
		},
		{
			name: "map value - permissive",
			input: map[string]any{
				"greeting": "Hello, {{name}}!",
				"info":     "{{name}} is {{age}} years old",
			},
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: map[string]any{
				"greeting": "Hello, John!",
				"info":     "John is 30 years old",
			},
			wantErr: false,
		},
		{
			name:  "array value - permissive",
			input: []any{"{{name}}", "{{age}}", "{{city}}"},
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: []any{"John", "30", "New York"},
			wantErr:  false,
		},
		{
			name: "nested structures - permissive",
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
			opts: RenderOptions{
				Mode: RenderModePermissive,
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
			wantErr: false,
		},
		{
			name:  "missing placeholder - strict mode",
			input: "Hello, {{missing}}!",
			opts: RenderOptions{
				Mode: RenderModeStrict,
			},
			wantErr: true,
		},
		{
			name: "map with missing placeholder - strict mode",
			input: map[string]any{
				"greeting": "Hello, {{name}}!",
				"info":     "Lives in {{missing}}",
			},
			opts: RenderOptions{
				Mode: RenderModeStrict,
			},
			wantErr: true,
		},
		{
			name:  "non-string value",
			input: 42,
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: 42,
			wantErr:  false,
		},
		{
			name:  "boolean value",
			input: true,
			opts: RenderOptions{
				Mode: RenderModePermissive,
			},
			expected: true,
			wantErr:  false,
		},
		{
			name:  "missing with default value",
			input: "Hello, {{missing}}!",
			opts: RenderOptions{
				Mode:         RenderModePermissive,
				DefaultValue: "World",
			},
			expected: "Hello, World!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderAnyWithOptions(tt.input, ctx, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenderAnyWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("RenderAnyWithOptions() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultRenderOptions(t *testing.T) {
	opts := DefaultRenderOptions()

	if opts.Mode != RenderModePermissive {
		t.Errorf("DefaultRenderOptions().Mode = %v, want %v", opts.Mode, RenderModePermissive)
	}

	if opts.DefaultValue != "" {
		t.Errorf("DefaultRenderOptions().DefaultValue = %q, want empty string", opts.DefaultValue)
	}
}
