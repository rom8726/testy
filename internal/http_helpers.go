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

	"github.com/kinbiko/jsonassert"
)

// ExecuteRequest performs an HTTP request against the given handler
func ExecuteRequest(t *testing.T, step Step, handler http.Handler, ctxMap map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	const op = "ExecuteRequest"

	step.Request = renderRequest(step.Request, ctxMap)

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

// AssertResponse checks if the HTTP response matches the expected response
func AssertResponse(
	t *testing.T,
	respRecorder *httptest.ResponseRecorder,
	expected ResponseSpec,
) {
	t.Helper()
	const op = "AssertResponse"

	if expected.Headers != nil {
		for k, v := range expected.Headers {
			recordedHeader := respRecorder.Header().Get(k)
			expectedHeader := v

			if recordedHeader != expectedHeader || !strings.Contains(recordedHeader, expectedHeader) {
				httpErr := NewError(ErrHTTP, op, "unexpected response header").
					WithContext("expected", v).
					WithContext("actual", respRecorder.Header().Get(k))
				t.Fatalf("%+v", httpErr)
			}
		}
	}

	body := respRecorder.Body.String()

	if respRecorder.Result().StatusCode != expected.Status {
		httpErr := NewError(ErrHTTP, op, "unexpected status code").
			WithContext("expected", expected.Status).
			WithContext("actual", respRecorder.Result().StatusCode).
			WithContext("!body", strings.ReplaceAll(body, "\n", " "))
		t.Fatalf("%+v", httpErr)
	}

	if expected.JSON != "" {
		ja := jsonassert.New(t)
		ja.Assert(body, expected.JSON)
		// Note: jsonassert.Assert calls t.Error internally if assertion fails
	} else if expected.Text != "" {
		if body != expected.Text {
			httpErr := NewError(ErrHTTP, op, "unexpected response body").
				WithContext("expected", expected.Text).
				WithContext("actual", body)
			t.Fatalf("%+v", httpErr)
		}
	}
}
