package internal

import (
	"reflect"
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	ctx := map[string]any{
		"age":       30,
		"name":      "John",
		"isPremium": true,
		"status":    "active",
		"empty":     "",
	}

	tests := []struct {
		name      string
		condition string
		want      bool
		wantErr   bool
	}{
		{
			name:      "empty condition",
			condition: "",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "equals string",
			condition: "{{name}} == John",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "not equals string",
			condition: "{{name}} != Alice",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "greater than number",
			condition: "{{age}} > 25",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "less than number",
			condition: "{{age}} < 50",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "greater or equal",
			condition: "{{age}} >= 30",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "less or equal",
			condition: "{{age}} <= 30",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "truthy check - true",
			condition: "{{isPremium}}",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "truthy check - false",
			condition: "{{empty}}",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "equals with quotes",
			condition: `{{status}} == "active"`,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "not equals with quotes",
			condition: `{{status}} != "inactive"`,
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCondition(tt.condition, ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareNumbers(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		operator string
		right    float64
		want     bool
		wantErr  bool
	}{
		{"equal", 5.0, "==", 5.0, true, false},
		{"not equal same", 5.0, "!=", 5.0, false, false},
		{"not equal different", 5.0, "!=", 10.0, true, false},
		{"greater than true", 10.0, ">", 5.0, true, false},
		{"greater than false", 5.0, ">", 10.0, false, false},
		{"less than true", 5.0, "<", 10.0, true, false},
		{"less than false", 10.0, "<", 5.0, false, false},
		{"greater or equal true", 10.0, ">=", 10.0, true, false},
		{"less or equal true", 5.0, "<=", 5.0, true, false},
		{"unknown operator", 5.0, "~=", 5.0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareNumbers(tt.left, tt.operator, tt.right)

			if (err != nil) != tt.wantErr {
				t.Errorf("compareNumbers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("compareNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareStrings(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		operator string
		right    string
		want     bool
		wantErr  bool
	}{
		{"equal", "hello", "==", "hello", true, false},
		{"not equal", "hello", "!=", "world", true, false},
		{"equal with quotes", `"hello"`, "==", `"hello"`, true, false},
		{"unsupported operator", "hello", ">", "world", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareStrings(tt.left, tt.operator, tt.right)

			if (err != nil) != tt.wantErr {
				t.Errorf("compareStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("compareStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", false},
		{"false", "false", false},
		{"False", "False", false},
		{"0", "0", false},
		{"null", "null", false},
		{"true", "true", true},
		{"True", "True", true},
		{"1", "1", true},
		{"any text", "hello", true},
		{"number", "42", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTruthy(tt.input); got != tt.want {
				t.Errorf("isTruthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldExecuteStep(t *testing.T) {
	ctx := map[string]any{
		"status": "active",
		"count":  5,
	}

	tests := []struct {
		name    string
		step    Step
		want    bool
		wantErr bool
	}{
		{
			name: "no condition",
			step: Step{
				Name: "test",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "condition true",
			step: Step{
				Name: "test",
				When: "{{status}} == active",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "condition false",
			step: Step{
				Name: "test",
				When: "{{status}} == inactive",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "numeric condition true",
			step: Step{
				Name: "test",
				When: "{{count}} > 3",
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ShouldExecuteStep(tt.step, ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ShouldExecuteStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ShouldExecuteStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandLoop_Items(t *testing.T) {
	ctx := map[string]any{
		"prefix": "user",
	}

	loop := &LoopConfig{
		Items: []any{"Alice", "Bob", "Charlie"},
		Var:   "name",
	}

	contexts, err := ExpandLoop(loop, ctx)
	if err != nil {
		t.Fatalf("ExpandLoop() error = %v", err)
	}

	if len(contexts) != 3 {
		t.Fatalf("Expected 3 contexts, got %d", len(contexts))
	}

	// Check first context
	if contexts[0]["name"] != "Alice" {
		t.Errorf("contexts[0][name] = %v, want Alice", contexts[0]["name"])
	}
	if contexts[0]["loopIndex"] != 0 {
		t.Errorf("contexts[0][loopIndex] = %v, want 0", contexts[0]["loopIndex"])
	}

	// Check second context
	if contexts[1]["name"] != "Bob" {
		t.Errorf("contexts[1][name] = %v, want Bob", contexts[1]["name"])
	}
	if contexts[1]["loopIndex"] != 1 {
		t.Errorf("contexts[1][loopIndex] = %v, want 1", contexts[1]["loopIndex"])
	}

	// Check that parent context is preserved
	if contexts[0]["prefix"] != "user" {
		t.Errorf("Parent context not preserved")
	}
}

func TestExpandLoop_Range(t *testing.T) {
	ctx := map[string]any{}

	tests := []struct {
		name      string
		loop      *LoopConfig
		wantCount int
		wantFirst int
		wantLast  int
	}{
		{
			name: "simple range",
			loop: &LoopConfig{
				Var: "i",
				Range: &Range{
					From: 1,
					To:   5,
					Step: 1,
				},
			},
			wantCount: 5,
			wantFirst: 1,
			wantLast:  5,
		},
		{
			name: "range with step 2",
			loop: &LoopConfig{
				Var: "i",
				Range: &Range{
					From: 0,
					To:   10,
					Step: 2,
				},
			},
			wantCount: 6,
			wantFirst: 0,
			wantLast:  10,
		},
		{
			name: "range with default step",
			loop: &LoopConfig{
				Var: "i",
				Range: &Range{
					From: 1,
					To:   3,
				},
			},
			wantCount: 3,
			wantFirst: 1,
			wantLast:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contexts, err := ExpandLoop(tt.loop, ctx)
			if err != nil {
				t.Fatalf("ExpandLoop() error = %v", err)
			}

			if len(contexts) != tt.wantCount {
				t.Errorf("Expected %d contexts, got %d", tt.wantCount, len(contexts))
			}

			if len(contexts) > 0 {
				if contexts[0]["i"] != tt.wantFirst {
					t.Errorf("First item = %v, want %v", contexts[0]["i"], tt.wantFirst)
				}

				if contexts[len(contexts)-1]["i"] != tt.wantLast {
					t.Errorf("Last item = %v, want %v", contexts[len(contexts)-1]["i"], tt.wantLast)
				}
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"age":  30,
		"address": map[string]any{
			"city":   "New York",
			"street": "5th Avenue",
		},
		"tags": []any{"admin", "user", "premium"},
		"orders": []any{
			map[string]any{"id": 1, "total": 100.0},
			map[string]any{"id": 2, "total": 200.0},
		},
	}

	tests := []struct {
		name    string
		path    string
		want    any
		wantErr bool
	}{
		{
			name:    "simple field",
			path:    "name",
			want:    "John",
			wantErr: false,
		},
		{
			name:    "nested field",
			path:    "address.city",
			want:    "New York",
			wantErr: false,
		},
		{
			name:    "array index",
			path:    "tags[0]",
			want:    "admin",
			wantErr: false,
		},
		{
			name:    "nested array",
			path:    "orders[1].total",
			want:    200.0,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			want:    data,
			wantErr: false,
		},
		{
			name:    "non-existent field",
			path:    "missing",
			wantErr: true,
		},
		{
			name:    "out of range index",
			path:    "tags[10]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSONPath(tt.path, data)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJSONPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePathParts(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []pathPart
	}{
		{
			name: "simple field",
			path: "name",
			want: []pathPart{{key: "name"}},
		},
		{
			name: "nested field",
			path: "user.name",
			want: []pathPart{{key: "user"}, {key: "name"}},
		},
		{
			name: "array index",
			path: "items[0]",
			want: []pathPart{{key: "items"}, {isArray: true, index: 0}},
		},
		{
			name: "complex path",
			path: "data.users[1].address.city",
			want: []pathPart{
				{key: "data"},
				{key: "users"},
				{isArray: true, index: 1},
				{key: "address"},
				{key: "city"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePathParts(tt.path)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePathParts() = %v, want %v", got, tt.want)
			}
		})
	}
}
