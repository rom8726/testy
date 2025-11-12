package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EvaluateCondition evaluates a conditional expression
// Supported formats:
// - "{{var}} == value"
// - "{{var}} != value"
// - "{{var}} > number"
// - "{{var}} < number"
// - "{{var}} >= number"
// - "{{var}} <= number"
// - "{{var}}" (truthy check)
func EvaluateCondition(condition string, ctx map[string]any) (bool, error) {
	if condition == "" {
		return true, nil
	}

	// First, render the condition to replace placeholders
	rendered := RenderTemplate(condition, ctx)

	// Parse operators
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}

	for _, op := range operators {
		if strings.Contains(rendered, op) {
			parts := strings.SplitN(rendered, op, 2)
			if len(parts) != 2 {
				return false, fmt.Errorf("invalid condition format: %s", condition)
			}

			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			return evaluateComparison(left, op, right)
		}
	}

	// No operator found - check if value is truthy
	return isTruthy(rendered), nil
}

// evaluateComparison compares two values using the given operator
func evaluateComparison(left, operator, right string) (bool, error) {
	// Try numeric comparison first
	leftNum, leftIsNum := parseNumber(left)
	rightNum, rightIsNum := parseNumber(right)

	if leftIsNum && rightIsNum {
		return compareNumbers(leftNum, operator, rightNum)
	}

	// String comparison
	return compareStrings(left, operator, right)
}

// compareNumbers compares two numbers
func compareNumbers(left float64, operator string, right float64) (bool, error) {
	switch operator {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case ">":
		return left > right, nil
	case "<":
		return left < right, nil
	case ">=":
		return left >= right, nil
	case "<=":
		return left <= right, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// compareStrings compares two strings
func compareStrings(left, operator, right string) (bool, error) {
	// Remove quotes if present
	left = strings.Trim(left, `"'`)
	right = strings.Trim(right, `"'`)

	switch operator {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	default:
		return false, fmt.Errorf("operator %s not supported for string comparison", operator)
	}
}

// parseNumber tries to parse a string as a number
func parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num, true
	}
	return 0, false
}

// isTruthy checks if a value is truthy
func isTruthy(s string) bool {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// Empty string is falsy
	if s == "" {
		return false
	}

	// Explicit false values
	falseValues := []string{"false", "0", "null", "nil", "undefined"}
	for _, fv := range falseValues {
		if s == fv {
			return false
		}
	}

	return true
}

// ShouldExecuteStep determines if a step should be executed based on its condition
func ShouldExecuteStep(step Step, ctx map[string]any) (bool, error) {
	if step.When == "" {
		return true, nil
	}

	return EvaluateCondition(step.When, ctx)
}

// ExpandLoop expands a loop configuration into individual items
func ExpandLoop(loop *LoopConfig, ctx map[string]any) ([]map[string]any, error) {
	if loop == nil {
		return nil, nil
	}

	var items []any

	// Use items list if provided
	if len(loop.Items) > 0 {
		// Render items to expand any placeholders
		for _, item := range loop.Items {
			rendered := RenderAny(item, ctx)
			items = append(items, rendered)
		}
	} else if loop.Range != nil {
		// Generate range
		step := loop.Range.Step
		if step == 0 {
			step = 1
		}

		if step > 0 {
			for i := loop.Range.From; i <= loop.Range.To; i += step {
				items = append(items, i)
			}
		} else {
			for i := loop.Range.From; i >= loop.Range.To; i += step {
				items = append(items, i)
			}
		}
	} else {
		return nil, fmt.Errorf("loop must have either items or range")
	}

	// Create context for each item
	var contexts []map[string]any
	for i, item := range items {
		itemCtx := make(map[string]any)
		// Copy parent context
		for k, v := range ctx {
			itemCtx[k] = v
		}
		// Add loop variables
		if loop.Var != "" {
			itemCtx[loop.Var] = item
		}
		itemCtx["loopIndex"] = i
		itemCtx["loopItem"] = item

		contexts = append(contexts, itemCtx)
	}

	return contexts, nil
}

// ParseJSONPath extracts a value from a map using a JSON path
// Supports: "field", "nested.field", "array[0]", "nested.array[0].field"
func ParseJSONPath(path string, data map[string]any) (any, error) {
	if path == "" {
		return data, nil
	}

	parts := parsePathParts(path)

	var current any = data

	for _, part := range parts {
		if part.isArray {
			// Array access
			arr, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("path %s: expected array", path)
			}

			if part.index < 0 || part.index >= len(arr) {
				return nil, fmt.Errorf("path %s: index %d out of range", path, part.index)
			}

			current = arr[part.index]
		} else {
			// Object access
			obj, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("path %s: expected object", path)
			}

			val, exists := obj[part.key]
			if !exists {
				return nil, fmt.Errorf("path %s: key %s not found", path, part.key)
			}

			current = val
		}
	}

	return current, nil
}

type pathPart struct {
	key     string
	isArray bool
	index   int
}

// parsePathParts parses a JSON path into parts
func parsePathParts(path string) []pathPart {
	var parts []pathPart

	// Regex to match parts: "field", "field[0]", "[0]"
	re := regexp.MustCompile(`([^.\[]+)|\[(\d+)\]`)
	matches := re.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		if match[1] != "" {
			// Object key
			parts = append(parts, pathPart{key: match[1]})
		} else if match[2] != "" {
			// Array index
			index, _ := strconv.Atoi(match[2])
			parts = append(parts, pathPart{isArray: true, index: index})
		}
	}

	return parts
}
