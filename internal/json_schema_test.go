package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateJSONSchema_BasicTypes(t *testing.T) {
	tests := []struct {
		name      string
		data      any
		schema    JSONSchema
		wantError bool
	}{
		{
			name: "valid string",
			data: "hello",
			schema: JSONSchema{
				Type: "string",
			},
			wantError: false,
		},
		{
			name: "invalid string type",
			data: 123,
			schema: JSONSchema{
				Type: "string",
			},
			wantError: true,
		},
		{
			name: "valid number",
			data: 42.5,
			schema: JSONSchema{
				Type: "number",
			},
			wantError: false,
		},
		{
			name: "valid integer",
			data: 42,
			schema: JSONSchema{
				Type: "integer",
			},
			wantError: false,
		},
		{
			name: "invalid integer (float)",
			data: 42.5,
			schema: JSONSchema{
				Type: "integer",
			},
			wantError: true,
		},
		{
			name: "valid boolean",
			data: true,
			schema: JSONSchema{
				Type: "boolean",
			},
			wantError: false,
		},
		{
			name: "valid array",
			data: []any{1, 2, 3},
			schema: JSONSchema{
				Type: "array",
			},
			wantError: false,
		},
		{
			name: "valid object",
			data: map[string]any{"key": "value"},
			schema: JSONSchema{
				Type: "object",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, tt.schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestValidateJSONSchema_ObjectProperties(t *testing.T) {
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"name": {
				Type: "string",
			},
			"age": {
				Type: "integer",
			},
			"email": {
				Type: "string",
			},
		},
		Required: []string{"name", "age"},
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
		errorPath string
	}{
		{
			name: "valid object",
			data: map[string]any{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
			wantError: false,
		},
		{
			name: "missing required field",
			data: map[string]any{
				"name": "John",
			},
			wantError: true,
			errorPath: "root/age",
		},
		{
			name: "wrong type for field",
			data: map[string]any{
				"name": "John",
				"age":  "thirty",
			},
			wantError: true,
			errorPath: "root/age",
		},
		{
			name: "valid with optional field missing",
			data: map[string]any{
				"name": "John",
				"age":  30,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}

			if tt.wantError && len(errors) > 0 && tt.errorPath != "" {
				found := false
				for _, err := range errors {
					if err.Path == tt.errorPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error at path %s, got errors: %v", tt.errorPath, errors)
				}
			}
		})
	}
}

func TestValidateJSONSchema_Arrays(t *testing.T) {
	schema := JSONSchema{
		Type: "array",
		Items: &JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"id": {
					Type: "integer",
				},
				"name": {
					Type: "string",
				},
			},
			Required: []string{"id"},
		},
	}

	tests := []struct {
		name      string
		data      []any
		wantError bool
	}{
		{
			name: "valid array",
			data: []any{
				map[string]any{"id": 1, "name": "Item 1"},
				map[string]any{"id": 2, "name": "Item 2"},
			},
			wantError: false,
		},
		{
			name: "invalid item in array",
			data: []any{
				map[string]any{"id": 1, "name": "Item 1"},
				map[string]any{"name": "Item 2"}, // missing id
			},
			wantError: true,
		},
		{
			name: "wrong type in array item",
			data: []any{
				map[string]any{"id": "one", "name": "Item 1"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestValidateJSONSchema_StringConstraints(t *testing.T) {
	minLen := 3
	maxLen := 10

	schema := JSONSchema{
		Type:      "string",
		MinLength: &minLen,
		MaxLength: &maxLen,
	}

	tests := []struct {
		name      string
		data      string
		wantError bool
	}{
		{
			name:      "valid string",
			data:      "hello",
			wantError: false,
		},
		{
			name:      "too short",
			data:      "hi",
			wantError: true,
		},
		{
			name:      "too long",
			data:      "this is way too long",
			wantError: true,
		},
		{
			name:      "minimum length",
			data:      "abc",
			wantError: false,
		},
		{
			name:      "maximum length",
			data:      "abcdefghij",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestValidateJSONSchema_NumberConstraints(t *testing.T) {
	min := 10.0
	max := 100.0

	schema := JSONSchema{
		Type:    "number",
		Minimum: &min,
		Maximum: &max,
	}

	tests := []struct {
		name      string
		data      float64
		wantError bool
	}{
		{
			name:      "valid number",
			data:      50.0,
			wantError: false,
		},
		{
			name:      "too small",
			data:      5.0,
			wantError: true,
		},
		{
			name:      "too large",
			data:      150.0,
			wantError: true,
		},
		{
			name:      "minimum value",
			data:      10.0,
			wantError: false,
		},
		{
			name:      "maximum value",
			data:      100.0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestValidateJSONSchema_Enum(t *testing.T) {
	schema := JSONSchema{
		Type: "string",
		Enum: []any{"red", "green", "blue"},
	}

	tests := []struct {
		name      string
		data      string
		wantError bool
	}{
		{
			name:      "valid enum value",
			data:      "red",
			wantError: false,
		},
		{
			name:      "another valid value",
			data:      "blue",
			wantError: false,
		},
		{
			name:      "invalid enum value",
			data:      "yellow",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestValidateJSONSchema_AdditionalProperties(t *testing.T) {
	falseValue := false
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"name": {Type: "string"},
		},
		AdditionalProperties: &falseValue,
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "no additional properties",
			data: map[string]any{
				"name": "John",
			},
			wantError: false,
		},
		{
			name: "with additional property",
			data: map[string]any{
				"name": "John",
				"age":  30,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateJSONSchema(tt.data, schema, "root")
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateJSONSchema() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestLoadJSONSchemaFromFile(t *testing.T) {
	// Create a temporary schema file
	tempDir, err := os.MkdirTemp("", "schema-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	schemaPath := filepath.Join(tempDir, "schema.json")
	schemaContent := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		},
		"required": ["name"]
	}`

	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Test loading
	schema, err := LoadJSONSchemaFromFile(schemaPath)
	if err != nil {
		t.Fatalf("LoadJSONSchemaFromFile() error = %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required field 'name', got %v", schema.Required)
	}
}

func TestSchemaValidationError_Error(t *testing.T) {
	err := SchemaValidationError{
		Path:    "root/user/age",
		Message: "expected integer",
	}

	expected := "validation error at root/user/age: expected integer"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestValidateJSONSchema_ComplexNested(t *testing.T) {
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]JSONSchema{
			"user": {
				Type: "object",
				Properties: map[string]JSONSchema{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
					"address": {
						Type: "object",
						Properties: map[string]JSONSchema{
							"city":   {Type: "string"},
							"street": {Type: "string"},
						},
						Required: []string{"city"},
					},
				},
				Required: []string{"name"},
			},
			"tags": {
				Type: "array",
				Items: &JSONSchema{
					Type: "string",
				},
			},
		},
	}

	data := map[string]any{
		"user": map[string]any{
			"name": "John",
			"age":  30,
			"address": map[string]any{
				"city":   "New York",
				"street": "5th Avenue",
			},
		},
		"tags": []any{"admin", "user"},
	}

	errors := ValidateJSONSchema(data, schema, "root")
	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid complex nested object, got: %v", errors)
	}

	// Test with missing required nested field
	invalidData := map[string]any{
		"user": map[string]any{
			"name": "John",
			"address": map[string]any{
				"street": "5th Avenue",
				// missing city
			},
		},
	}

	errors = ValidateJSONSchema(invalidData, schema, "root")
	if len(errors) == 0 {
		t.Error("Expected validation error for missing required nested field")
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantValue float64
		wantOk    bool
	}{
		{"float64", float64(42.5), 42.5, true},
		{"float32", float32(42.5), 42.5, true},
		{"int", int(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"string", "42", 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := toFloat64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("toFloat64() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && value != tt.wantValue {
				t.Errorf("toFloat64() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}
