package runner

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

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

		assert.Equal(t, tc.Response.Status, rec.Code)

		if tc.Response.JSON != nil {
			var actual map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
				t.Fatalf("invalid JSON response: %v", err)
			}

			for k, v := range tc.Response.JSON {
				assert.Equal(t, v, actual[k], "mismatch on key '%s'", k)
			}
		}
	})
}
