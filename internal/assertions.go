package internal

import (
	"fmt"
	"regexp"
	"strings"
)

func AssertResponseV2(assertion ResponseAssertion, responseData map[string]any) error {
	// Extract value at path
	value, err := ParseJSONPath(assertion.Path, responseData)
	if err != nil {
		if assertion.Message != "" {
			return fmt.Errorf("%s: %w", assertion.Message, err)
		}
		return err
	}

	// Evaluate assertion based on operator
	result, err := evaluateAssertion(value, assertion.Operator, assertion.Value)
	if err != nil {
		if assertion.Message != "" {
			return fmt.Errorf("%s: %w", assertion.Message, err)
		}
		return err
	}

	if !result {
		msg := "assertion failed at path %s: expected %s %v, got %v"

		if assertion.Message != "" {
			msg = assertion.Message + ": " + msg
		}

		return fmt.Errorf(msg, assertion.Path, assertion.Operator, assertion.Value, value)
	}

	return nil
}

// evaluateAssertion evaluates an assertion based on the operator
func evaluateAssertion(actual any, operator string, expected any) (bool, error) {
	switch operator {
	case "equals", "eq", "==":
		return assertEquals(actual, expected), nil

	case "notEquals", "ne", "!=":
		return !assertEquals(actual, expected), nil

	case "greaterThan", "gt", ">":
		return assertGreaterThan(actual, expected)

	case "lessThan", "lt", "<":
		return assertLessThan(actual, expected)

	case "greaterOrEqual", "gte", ">=":
		return assertGreaterOrEqual(actual, expected)

	case "lessOrEqual", "lte", "<=":
		return assertLessOrEqual(actual, expected)

	case "contains":
		return assertContains(actual, expected), nil

	case "notContains":
		return !assertContains(actual, expected), nil

	case "matches":
		return assertMatches(actual, expected)

	case "startsWith":
		return assertStartsWith(actual, expected), nil

	case "endsWith":
		return assertEndsWith(actual, expected), nil

	case "between":
		return assertBetween(actual, expected)

	case "in":
		return assertIn(actual, expected), nil

	case "notIn":
		return !assertIn(actual, expected), nil

	case "isEmpty":
		return assertIsEmpty(actual), nil

	case "isNotEmpty":
		return !assertIsEmpty(actual), nil

	case "hasLength":
		return assertHasLength(actual, expected)

	case "hasMinLength":
		return assertHasMinLength(actual, expected)

	case "hasMaxLength":
		return assertHasMaxLength(actual, expected)

	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// assertEquals checks if two values are equal
func assertEquals(actual, expected any) bool {
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// assertGreaterThan checks if actual > expected
func assertGreaterThan(actual, expected any) (bool, error) {
	actualNum, actualOk := toFloat64(actual)
	expectedNum, expectedOk := toFloat64(expected)

	if !actualOk || !expectedOk {
		return false, fmt.Errorf("cannot compare non-numeric values with >")
	}

	return actualNum > expectedNum, nil
}

// assertLessThan checks if actual < expected
func assertLessThan(actual, expected any) (bool, error) {
	actualNum, actualOk := toFloat64(actual)
	expectedNum, expectedOk := toFloat64(expected)

	if !actualOk || !expectedOk {
		return false, fmt.Errorf("cannot compare non-numeric values with <")
	}

	return actualNum < expectedNum, nil
}

// assertGreaterOrEqual checks if actual >= expected
func assertGreaterOrEqual(actual, expected any) (bool, error) {
	actualNum, actualOk := toFloat64(actual)
	expectedNum, expectedOk := toFloat64(expected)

	if !actualOk || !expectedOk {
		return false, fmt.Errorf("cannot compare non-numeric values with >=")
	}

	return actualNum >= expectedNum, nil
}

// assertLessOrEqual checks if actual <= expected
func assertLessOrEqual(actual, expected any) (bool, error) {
	actualNum, actualOk := toFloat64(actual)
	expectedNum, expectedOk := toFloat64(expected)

	if !actualOk || !expectedOk {
		return false, fmt.Errorf("cannot compare non-numeric values with <=")
	}

	return actualNum <= expectedNum, nil
}

// assertContains checks if actual contains expected
func assertContains(actual, expected any) bool {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)

	// For arrays, check if element exists
	if arr, ok := actual.([]any); ok {
		for _, item := range arr {
			if fmt.Sprintf("%v", item) == expectedStr {
				return true
			}
		}
		return false
	}

	// For strings, check substring
	return strings.Contains(actualStr, expectedStr)
}

// assertMatches checks if actual matches expected regex pattern
func assertMatches(actual, expected any) (bool, error) {
	actualStr := fmt.Sprintf("%v", actual)
	patternStr := fmt.Sprintf("%v", expected)

	matched, err := regexp.MatchString(patternStr, actualStr)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return matched, nil
}

// assertStartsWith checks if actual starts with expected
func assertStartsWith(actual, expected any) bool {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)

	return strings.HasPrefix(actualStr, expectedStr)
}

// assertEndsWith checks if actual ends with expected
func assertEndsWith(actual, expected any) bool {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)

	return strings.HasSuffix(actualStr, expectedStr)
}

// assertBetween checks if actual is between min and max (expected should be [min, max])
func assertBetween(actual, expected any) (bool, error) {
	actualNum, actualOk := toFloat64(actual)
	if !actualOk {
		return false, fmt.Errorf("actual value must be numeric for between")
	}

	// Expected should be an array with [min, max]
	expectedArr, ok := expected.([]any)
	if !ok || len(expectedArr) != 2 {
		return false, fmt.Errorf("expected value for 'between' must be array [min, max]")
	}

	minNum, minOk := toFloat64(expectedArr[0])
	maxNum, maxOk := toFloat64(expectedArr[1])

	if !minOk || !maxOk {
		return false, fmt.Errorf("min and max values must be numeric")
	}

	return actualNum >= minNum && actualNum <= maxNum, nil
}

// assertIn checks if actual is in expected list
func assertIn(actual, expected any) bool {
	expectedArr, ok := expected.([]any)
	if !ok {
		return false
	}

	actualStr := fmt.Sprintf("%v", actual)

	for _, item := range expectedArr {
		if fmt.Sprintf("%v", item) == actualStr {
			return true
		}
	}

	return false
}

// assertIsEmpty checks if actual is empty
func assertIsEmpty(actual any) bool {
	switch v := actual.(type) {
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	case nil:
		return true
	}

	return false
}

// assertHasLength checks if actual has specific length
func assertHasLength(actual, expected any) (bool, error) {
	expectedNum, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("expected value for hasLength must be numeric")
	}

	length := getLength(actual)
	if length < 0 {
		return false, fmt.Errorf("cannot get length of non-collection type")
	}

	return float64(length) == expectedNum, nil
}

// assertHasMinLength checks if actual has minimum length
func assertHasMinLength(actual, expected any) (bool, error) {
	expectedNum, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("expected value for hasMinLength must be numeric")
	}

	length := getLength(actual)
	if length < 0 {
		return false, fmt.Errorf("cannot get length of non-collection type")
	}

	return float64(length) >= expectedNum, nil
}

// assertHasMaxLength checks if actual has maximum length
func assertHasMaxLength(actual, expected any) (bool, error) {
	expectedNum, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("expected value for hasMaxLength must be numeric")
	}

	length := getLength(actual)
	if length < 0 {
		return false, fmt.Errorf("cannot get length of non-collection type")
	}

	return float64(length) <= expectedNum, nil
}

// getLength returns the length of a collection or -1 if not applicable
func getLength(v any) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []any:
		return len(val)
	case map[string]any:
		return len(val)
	}

	return -1
}
