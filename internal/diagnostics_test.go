package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormatHTTPDiagnosticIncludesRenderedRequestAndResponse(t *testing.T) {
	step := Step{
		Name: "get_user",
		Request: &RequestSpec{
			Method: "GET",
			Path:   "/users/{{userID}}",
			Headers: map[string]string{
				"Authorization": "Bearer {{token}}",
			},
			Body: map[string]any{"id": "{{userID}}"},
		},
	}
	step = renderStepForDiagnostics(step, map[string]any{
		"userID": 42,
		"token":  "abc",
	})

	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusNotFound)
	_, _ = rec.WriteString(`{"error":"missing"}`)

	got := formatHTTPDiagnostic(step, rec)
	for _, want := range []string{
		"step: get_user",
		"request: GET /users/42",
		`Authorization="Bearer abc"`,
		`request body: {"id":42}`,
		"response status: 404",
		`Content-Type="application/json"`,
		`response body: {"error":"missing"}`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatHTTPDiagnostic() should contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatDBDiagnostic(t *testing.T) {
	check := DBCheck{
		Query:  "SELECT id FROM users",
		Result: []map[string]any{{"id": 1}},
	}

	got := formatDBDiagnostic("check_user", check, `[{"id":2}]`)
	for _, want := range []string{
		"step: check_user",
		"query: SELECT id FROM users",
		`expected rows: [{"id":1}]`,
		`actual rows: [{"id":2}]`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatDBDiagnostic() should contain %q, got:\n%s", want, got)
		}
	}
}

func TestTruncateDiagnosticValue(t *testing.T) {
	input := strings.Repeat("x", diagnosticValueLimit+1)
	got := truncateDiagnosticValue(input)

	if len(got) <= diagnosticValueLimit {
		t.Fatalf("truncateDiagnosticValue() did not append truncation marker")
	}
	if !strings.Contains(got, "truncated") {
		t.Fatalf("truncateDiagnosticValue() should include truncation marker, got %q", got)
	}
}
