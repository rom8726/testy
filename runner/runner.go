package runner

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinbiko/jsonassert"

	"github.com/rom8726/testy/types"
)

func RunSingle(t *testing.T, handler http.Handler, tc types.TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		var body io.Reader
		if tc.Request.Body != nil {
			b, _ := json.Marshal(tc.Request.Body)
			body = bytes.NewReader(b)
		}

		req := httptest.NewRequest(tc.Request.Method, tc.Request.Path, body)
		for k, v := range tc.Request.Headers {
			req.Header.Set(k, v)
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertResponse(t, rec.Result(), rec, tc.Response)
	})
}

func assertResponse(
	t *testing.T,
	resp *http.Response,
	respRecorder *httptest.ResponseRecorder,
	expected types.ResponseSpec,
) {
	if resp.StatusCode != expected.Status {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, expected.Status)
	}

	body := respRecorder.Body.String()

	if expected.JSON != "" {
		ja := jsonassert.New(t)
		ja.Assertf(body, expected.JSON)
	}
}
