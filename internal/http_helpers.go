package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// ExecuteRequest performs an HTTP request with full rendering support
func ExecuteRequest(t *testing.T, step Step, handler http.Handler, ctxMap map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	const op = "ExecuteRequest"

	// Render request with context and faker
	renderedReq := renderRequest(*step.Request, ctxMap)
	step.Request = &renderedReq

	var body io.Reader
	if step.Request.BodyFile != "" {
		bodyData, err := os.ReadFile(step.Request.BodyFile)
		if err != nil {
			httpErr := NewError(ErrHTTP, op, "failed to read request body file").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", httpErr)
		}
		body = bytes.NewReader(bodyData)
	} else if step.Request.BodyRaw != "" {
		body = strings.NewReader(step.Request.BodyRaw)
	} else if step.Request.Body != nil {
		b, err := json.Marshal(step.Request.Body)
		if err != nil {
			httpErr := NewError(ErrHTTP, op, "failed to marshal request body").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", httpErr)
		}
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

// AssertResponse checks HTTP response with legacy JSON assertion.
func AssertResponse(
	t *testing.T,
	respRecorder *httptest.ResponseRecorder,
	expected ResponseSpec,
) {
	t.Helper()
	AssertStepResponse(t, Step{}, respRecorder, expected)
}

// AssertStepResponse checks HTTP response and includes step/request context in failure diagnostics.
func AssertStepResponse(
	t *testing.T,
	step Step,
	respRecorder *httptest.ResponseRecorder,
	expected ResponseSpec,
) {
	t.Helper()
	const op = "AssertResponse"

	// Check headers
	if expected.Headers != nil {
		for k, v := range expected.Headers {
			recordedHeader := respRecorder.Header().Get(k)
			expectedHeader := v

			if recordedHeader != expectedHeader &&
				!strings.Contains(recordedHeader, expectedHeader) &&
				!strings.Contains(expectedHeader, recordedHeader) {
				httpErr := NewError(ErrHTTP, op, "unexpected response header").
					WithContext("step", step.Name).
					WithContext("expected", v).
					WithContext("actual", respRecorder.Header().Get(k)).
					WithContext("!diagnostic", formatHTTPDiagnostic(step, respRecorder))
				t.Fatalf("%+v", httpErr)
			}
		}
	}

	body := respRecorder.Body.String()

	// Check status code
	if respRecorder.Result().StatusCode != expected.Status {
		httpErr := NewError(ErrHTTP, op, "unexpected status code").
			WithContext("step", step.Name).
			WithContext("expected", expected.Status).
			WithContext("actual", respRecorder.Result().StatusCode).
			WithContext("!diagnostic", formatHTTPDiagnostic(step, respRecorder))
		t.Fatalf("%+v", httpErr)
	}

	// Check JSON (legacy format)
	if expected.JSON != "" {
		assertJSONWithDiagnostic(
			t,
			body,
			expected.JSON,
			"HTTP response assertion context",
			formatHTTPDiagnostic(step, respRecorder),
		)
	} else if expected.Text != "" {
		if body != expected.Text {
			httpErr := NewError(ErrHTTP, op, "unexpected response body").
				WithContext("step", step.Name).
				WithContext("expected", expected.Text).
				WithContext("actual", body).
				WithContext("!diagnostic", formatHTTPDiagnostic(step, respRecorder))
			t.Fatalf("%+v", httpErr)
		}
	}
}
