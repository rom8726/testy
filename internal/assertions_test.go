package internal

import (
	"testing"
)

func TestAssertEquals(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"equal strings", "hello", "hello", true},
		{"unequal strings", "hello", "world", false},
		{"equal numbers", 42, 42, true},
		{"unequal numbers", 42, 43, false},
		{"equal booleans", true, true, true},
		{"string number match", "42", 42, true}, // Converted to strings
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assertEquals(tt.actual, tt.expected); got != tt.want {
				t.Errorf("assertEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertGreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
		wantErr  bool
	}{
		{"greater", 10, 5, true, false},
		{"not greater", 5, 10, false, false},
		{"equal", 5, 5, false, false},
		{"non-numeric", "hello", 5, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertGreaterThan(tt.actual, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("assertGreaterThan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("assertGreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertContains(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"string contains", "hello world", "world", true},
		{"string not contains", "hello world", "foo", false},
		{"array contains", []any{1, 2, 3}, 2, true},
		{"array not contains", []any{1, 2, 3}, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assertContains(tt.actual, tt.expected); got != tt.want {
				t.Errorf("assertContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertMatches(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
		wantErr  bool
	}{
		{"valid regex match", "hello123", `^[a-z]+\d+$`, true, false},
		{"valid regex no match", "hello", `^\d+$`, false, false},
		{"email pattern", "test@example.com", `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`, true, false},
		{"invalid regex", "test", `[`, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertMatches(tt.actual, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("assertMatches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("assertMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertStartsWith(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"starts with", "hello world", "hello", true},
		{"not starts with", "hello world", "world", false},
		{"exact match", "hello", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assertStartsWith(tt.actual, tt.expected); got != tt.want {
				t.Errorf("assertStartsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertEndsWith(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"ends with", "hello world", "world", true},
		{"not ends with", "hello world", "hello", false},
		{"exact match", "hello", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assertEndsWith(tt.actual, tt.expected); got != tt.want {
				t.Errorf("assertEndsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertBetween(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
		wantErr  bool
	}{
		{"between", 50, []any{10, 100}, true, false},
		{"below range", 5, []any{10, 100}, false, false},
		{"above range", 150, []any{10, 100}, false, false},
		{"at min", 10, []any{10, 100}, true, false},
		{"at max", 100, []any{10, 100}, true, false},
		{"invalid range", 50, []any{10}, false, true},
		{"non-numeric actual", "hello", []any{10, 100}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertBetween(tt.actual, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("assertBetween() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("assertBetween() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertIn(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"in list", "apple", []any{"apple", "banana", "orange"}, true},
		{"not in list", "grape", []any{"apple", "banana", "orange"}, false},
		{"number in list", 2, []any{1, 2, 3}, true},
		{"invalid expected type", "apple", "not-an-array", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assertIn(tt.actual, tt.expected); got != tt.want {
				t.Errorf("assertIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		actual any
		want   bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"empty array", []any{}, true},
		{"non-empty array", []any{1, 2, 3}, false},
		{"empty map", map[string]any{}, true},
		{"non-empty map", map[string]any{"key": "value"}, false},
		{"nil", nil, true},
		{"number", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assertIsEmpty(tt.actual); got != tt.want {
				t.Errorf("assertIsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertHasLength(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
		wantErr  bool
	}{
		{"string length", "hello", 5, true, false},
		{"array length", []any{1, 2, 3}, 3, true, false},
		{"map length", map[string]any{"a": 1, "b": 2}, 2, true, false},
		{"wrong length", "hello", 10, false, false},
		{"non-collection", 42, 1, false, true},
		{"non-numeric expected", "hello", "five", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertHasLength(tt.actual, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("assertHasLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("assertHasLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertHasMinLength(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
		wantErr  bool
	}{
		{"above min", "hello world", 5, true, false},
		{"at min", "hello", 5, true, false},
		{"below min", "hi", 5, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertHasMinLength(tt.actual, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("assertHasMinLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("assertHasMinLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertHasMaxLength(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
		wantErr  bool
	}{
		{"below max", "hi", 10, true, false},
		{"at max", "hello", 5, true, false},
		{"above max", "hello world", 5, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertHasMaxLength(tt.actual, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("assertHasMaxLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("assertHasMaxLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateAssertion(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		operator string
		expected any
		want     bool
		wantErr  bool
	}{
		{"equals", 42, "equals", 42, true, false},
		{"eq", 42, "eq", 42, true, false},
		{"==", 42, "==", 42, true, false},
		{"notEquals", 42, "notEquals", 43, true, false},
		{"greaterThan", 10, "greaterThan", 5, true, false},
		{"lessThan", 5, "lessThan", 10, true, false},
		{"contains", "hello world", "contains", "world", true, false},
		{"matches", "test123", "matches", `\d+`, true, false},
		{"between", 50, "between", []any{10, 100}, true, false},
		{"in", "apple", "in", []any{"apple", "banana"}, true, false},
		{"isEmpty", "", "isEmpty", nil, true, false},
		{"hasLength", "hello", "hasLength", 5, true, false},
		{"unknown operator", 42, "unknown", 42, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateAssertion(tt.actual, tt.operator, tt.expected)

			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateAssertion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("evaluateAssertion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssertResponseV2(t *testing.T) {
	responseData := map[string]any{
		"status": "active",
		"count":  42,
		"tags":   []any{"admin", "user"},
		"user": map[string]any{
			"name": "John",
			"age":  30,
		},
	}

	tests := []struct {
		name      string
		assertion ResponseAssertion
		wantErr   bool
	}{
		{
			name: "valid equals assertion",
			assertion: ResponseAssertion{
				Path:     "status",
				Operator: "equals",
				Value:    "active",
			},
			wantErr: false,
		},
		{
			name: "valid greater than assertion",
			assertion: ResponseAssertion{
				Path:     "count",
				Operator: "greaterThan",
				Value:    40,
			},
			wantErr: false,
		},
		{
			name: "valid contains assertion",
			assertion: ResponseAssertion{
				Path:     "tags",
				Operator: "contains",
				Value:    "admin",
			},
			wantErr: false,
		},
		{
			name: "nested path assertion",
			assertion: ResponseAssertion{
				Path:     "user.name",
				Operator: "equals",
				Value:    "John",
			},
			wantErr: false,
		},
		{
			name: "failed assertion",
			assertion: ResponseAssertion{
				Path:     "count",
				Operator: "equals",
				Value:    100,
			},
			wantErr: true,
		},
		{
			name: "invalid path",
			assertion: ResponseAssertion{
				Path:     "nonexistent",
				Operator: "equals",
				Value:    "something",
			},
			wantErr: true,
		},
		{
			name: "custom error message",
			assertion: ResponseAssertion{
				Path:     "count",
				Operator: "equals",
				Value:    100,
				Message:  "Count should be 100",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertResponseV2(tt.assertion, responseData)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssertResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetLength(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{"string", "hello", 5},
		{"array", []any{1, 2, 3}, 3},
		{"map", map[string]any{"a": 1, "b": 2}, 2},
		{"number", 42, -1},
		{"bool", true, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLength(tt.value); got != tt.want {
				t.Errorf("getLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
