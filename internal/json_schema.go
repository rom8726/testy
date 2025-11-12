package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONSchema represents a JSON Schema for response validation
type JSONSchema struct {
	Type                 string                `json:"type,omitempty"`
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *JSONSchema           `json:"items,omitempty"`
	AdditionalProperties *bool                 `json:"additionalProperties,omitempty"`
	Enum                 []any                 `json:"enum,omitempty"`
	Minimum              *float64              `json:"minimum,omitempty"`
	Maximum              *float64              `json:"maximum,omitempty"`
	MinLength            *int                  `json:"minLength,omitempty"`
	MaxLength            *int                  `json:"maxLength,omitempty"`
	Pattern              string                `json:"pattern,omitempty"`
	Format               string                `json:"format,omitempty"`
}

// SchemaValidationError represents a schema validation error
type SchemaValidationError struct {
	Path    string
	Message string
}

func (e SchemaValidationError) Error() string {
	return fmt.Sprintf("validation error at %s: %s", e.Path, e.Message)
}

// ValidateJSONSchema validates a JSON value against a schema
func ValidateJSONSchema(data any, schema JSONSchema, path string) []SchemaValidationError {
	var errors []SchemaValidationError

	// Type validation
	if schema.Type != "" {
		if !validateType(data, schema.Type) {
			errors = append(errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("expected type %s, got %T", schema.Type, data),
			})

			return errors // Stop validation if the type is wrong
		}
	}

	// Enum validation
	if len(schema.Enum) > 0 {
		if !validateEnum(data, schema.Enum) {
			errors = append(errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("value must be one of %v", schema.Enum),
			})
		}
	}

	// Type-specific validations
	switch schema.Type {
	case "object":
		if objData, ok := data.(map[string]any); ok {
			errors = append(errors, validateObject(objData, schema, path)...)
		}
	case "array":
		if arrData, ok := data.([]any); ok {
			errors = append(errors, validateArray(arrData, schema, path)...)
		}
	case "string":
		if strData, ok := data.(string); ok {
			errors = append(errors, validateString(strData, schema, path)...)
		}
	case "number", "integer":
		if numData, ok := toFloat64(data); ok {
			errors = append(errors, validateNumber(numData, schema, path)...)
		}
	}

	return errors
}

// LoadJSONSchemaFromFile loads a JSON Schema from a file
func LoadJSONSchemaFromFile(path string) (JSONSchema, error) {
	var schema JSONSchema

	data, err := os.ReadFile(path)
	if err != nil {
		return schema, fmt.Errorf("failed to read schema file: %w", err)
	}

	if err := json.Unmarshal(data, &schema); err != nil {
		return schema, fmt.Errorf("failed to parse schema: %w", err)
	}

	return schema, nil
}

// validateType checks if the data matches the expected type
func validateType(data any, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := data.(string)
		return ok
	case "number":
		_, ok := toFloat64(data)
		return ok
	case "integer":
		if num, ok := toFloat64(data); ok {
			return num == float64(int64(num))
		}
		return false
	case "boolean":
		_, ok := data.(bool)
		return ok
	case "array":
		_, ok := data.([]any)
		return ok
	case "object":
		_, ok := data.(map[string]any)
		return ok
	case "null":
		return data == nil
	}
	return true
}

// validateEnum checks if value is in enum list
func validateEnum(data any, enum []any) bool {
	for _, e := range enum {
		if fmt.Sprintf("%v", data) == fmt.Sprintf("%v", e) {
			return true
		}
	}
	return false
}

// validateObject validates an object against schema
func validateObject(data map[string]any, schema JSONSchema, path string) []SchemaValidationError {
	var errors []SchemaValidationError

	// Required fields validation
	for _, required := range schema.Required {
		if _, exists := data[required]; !exists {
			errors = append(errors, SchemaValidationError{
				Path:    filepath.Join(path, required),
				Message: "required field is missing",
			})
		}
	}

	// Properties validation
	for key, value := range data {
		propSchema, hasPropSchema := schema.Properties[key]

		if !hasPropSchema {
			// Check additionalProperties
			if schema.AdditionalProperties != nil && !*schema.AdditionalProperties {
				errors = append(errors, SchemaValidationError{
					Path:    filepath.Join(path, key),
					Message: "additional property not allowed",
				})
			}
			continue
		}

		propPath := filepath.Join(path, key)
		errors = append(errors, ValidateJSONSchema(value, propSchema, propPath)...)
	}

	return errors
}

// validateArray validates an array against schema
func validateArray(data []any, schema JSONSchema, path string) []SchemaValidationError {
	var errors []SchemaValidationError

	if schema.Items != nil {
		for i, item := range data {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			errors = append(errors, ValidateJSONSchema(item, *schema.Items, itemPath)...)
		}
	}

	return errors
}

// validateString validates a string against schema
func validateString(data string, schema JSONSchema, path string) []SchemaValidationError {
	var errors []SchemaValidationError

	if schema.MinLength != nil && len(data) < *schema.MinLength {
		errors = append(errors, SchemaValidationError{
			Path:    path,
			Message: fmt.Sprintf("string length %d is less than minimum %d", len(data), *schema.MinLength),
		})
	}

	if schema.MaxLength != nil && len(data) > *schema.MaxLength {
		errors = append(errors, SchemaValidationError{
			Path:    path,
			Message: fmt.Sprintf("string length %d is greater than maximum %d", len(data), *schema.MaxLength),
		})
	}

	if schema.Pattern != "" {
		// Pattern validation would require regexp package
		// Left as TODO for now
	}

	return errors
}

// validateNumber validates a number against schema
func validateNumber(data float64, schema JSONSchema, path string) []SchemaValidationError {
	var errors []SchemaValidationError

	if schema.Minimum != nil && data < *schema.Minimum {
		errors = append(errors, SchemaValidationError{
			Path:    path,
			Message: fmt.Sprintf("value %f is less than minimum %f", data, *schema.Minimum),
		})
	}

	if schema.Maximum != nil && data > *schema.Maximum {
		errors = append(errors, SchemaValidationError{
			Path:    path,
			Message: fmt.Sprintf("value %f is greater than maximum %f", data, *schema.Maximum),
		})
	}

	return errors
}

// toFloat64 converts various numeric types to float64
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	}
	return 0, false
}
