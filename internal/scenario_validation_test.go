package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateTestCasesValidHTTPScenario(t *testing.T) {
	tempDir := t.TempDir()
	fixturesDir := filepath.Join(tempDir, "fixtures")
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatalf("failed to create fixtures dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fixturesDir, "users.yml"), []byte("users: []"), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	schemaPath := filepath.Join(tempDir, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"type":"object"}`), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	cases := []TestCase{
		{
			Name:      "user flow",
			Variables: map[string]any{"token": "abc"},
			Fixtures:  []string{"users"},
			Steps: []Step{
				{
					Name: "create_user",
					Request: &RequestSpec{
						Method: "POST",
						Path:   "/users",
						Body: map[string]any{
							"name":  "{{faker.name}}",
							"token": "{{token}}",
						},
					},
					Response: &ResponseSpec{
						Status: 201,
						Schema: schemaPath,
						Assertions: []ResponseAssertion{
							{Path: "id", Operator: "eq", Value: "{{create_user.response.id}}"},
						},
					},
				},
				{
					Name: "get user",
					Request: &RequestSpec{
						Method: "GET",
						Path:   "/users/{{create_user.response.id}}",
					},
					Response: &ResponseSpec{Status: 200},
				},
			},
		},
	}

	err := ValidateTestCases(cases, ValidationConfig{
		FixturesDir: fixturesDir,
		HasHandler:  true,
	})
	if err != nil {
		t.Fatalf("ValidateTestCases() unexpected error: %v", err)
	}
}

func TestValidateTestCasesReportsMultipleProblems(t *testing.T) {
	cases := []TestCase{
		{
			Name: "broken",
			MockCalls: []MockCallCheck{
				{Mock: "missing_mock", Count: -1},
			},
			Steps: []Step{
				{
					Name:     "dup name",
					Request:  &RequestSpec{Path: "/users/{{future.response.id}}"},
					Response: &ResponseSpec{Status: 99},
					Retry:    &RetryConfig{Attempts: 0, Backoff: "weird"},
				},
				{
					Name:     "dup name",
					Request:  &RequestSpec{Method: "GET", Path: "/users/{{missing.response.id}}"},
					Response: &ResponseSpec{Status: 200, Assertions: []ResponseAssertion{{Path: "id", Operator: "wat"}}},
				},
				{
					Name: "future",
					GRPCRequest: &GRPCRequestSpec{
						Service: "users.UserService",
						Method:  "GetUser",
					},
					GRPCResponse: &GRPCResponseSpec{Code: "NOPE"},
				},
			},
		},
	}

	err := ValidateTestCases(cases, ValidationConfig{})
	if err == nil {
		t.Fatal("ValidateTestCases() expected error, got nil")
	}

	msg := err.Error()
	for _, want := range []string{
		"duplicate step name",
		"HTTP request requires Config.Handler",
		"method is required",
		"status code 99 is outside 100..599",
		"unknown backoff strategy",
		"unknown operator \"wat\"",
		"references step \"future\" before it is available",
		"references unknown step or variable \"missing\"",
		"gRPC request requires Config.GRPCAddr",
		"grpcRequest: message is required",
		"unknown gRPC status code \"NOPE\"",
		"mock \"missing_mock\" is neither declared",
		"count cannot be negative",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("ValidateTestCases() error should contain %q, got %v", want, msg)
		}
	}
}

func TestValidateTestCasesValidatesMockManager(t *testing.T) {
	cases := []TestCase{
		{
			Name: "mocked flow",
			Mocks: map[string]MockServerDef{
				"notifications": {
					Routes: []MockRoute{
						{
							Method:   "POST",
							Path:     "/send",
							Response: MockResponse{Status: 202},
						},
					},
				},
			},
			MockCalls: []MockCallCheck{
				{Mock: "notifications", Count: 1},
			},
			Steps: []Step{
				{
					Name:     "ping",
					Request:  &RequestSpec{Method: "GET", Path: "/ping"},
					Response: &ResponseSpec{Status: 200},
				},
			},
		},
	}

	err := ValidateTestCases(cases, ValidationConfig{
		HasHandler: true,
		Mocks:      []*MockInstance{NewMockInstance("notifications", "http://mock", nil)},
	})
	if err != nil {
		t.Fatalf("ValidateTestCases() unexpected error: %v", err)
	}
}

func TestValidateTestCasesValidatesLoopPlaceholders(t *testing.T) {
	cases := []TestCase{
		{
			Name: "loop",
			Steps: []Step{
				{
					Name: "create",
					Loop: &LoopConfig{
						Range: &Range{From: 1, To: 3, Step: 1},
						Var:   "index",
					},
					Request: &RequestSpec{
						Method: "POST",
						Path:   "/users/{{index}}",
					},
					Response: &ResponseSpec{Status: 201},
				},
			},
		},
	}

	err := ValidateTestCases(cases, ValidationConfig{HasHandler: true})
	if err != nil {
		t.Fatalf("ValidateTestCases() unexpected error: %v", err)
	}
}
