package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinbiko/jsonassert"
)

// ExecuteRequest performs an HTTP request against the given handler
func ExecuteRequest(t *testing.T, step Step, handler http.Handler, ctxMap map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	step.Request = renderRequest(step.Request, ctxMap)

	var body io.Reader
	if step.Request.Body != nil {
		b, _ := json.Marshal(step.Request.Body)
		body = bytes.NewReader(b)
	}

	req := httptest.NewRequest(step.Request.Method, step.Request.Path, body)
	for k, v := range step.Request.Headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

// AssertResponse checks if the HTTP response matches the expected response
func AssertResponse(
	t *testing.T,
	respRecorder *httptest.ResponseRecorder,
	expected ResponseSpec,
) {
	t.Helper()

	if respRecorder.Result().StatusCode != expected.Status {
		t.Fatalf("unexpected status: got %d, want %d", respRecorder.Result().StatusCode, expected.Status)
	}

	body := respRecorder.Body.String()

	if expected.JSON != "" {
		ja := jsonassert.New(t)
		ja.Assert(body, expected.JSON)
	}
}
