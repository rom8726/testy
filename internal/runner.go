package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/rom8726/pgfixtures"
)

func RunSingle(t *testing.T, handler http.Handler, tc TestCase, cfg *Config) {
	t.Run(tc.Name, func(t *testing.T) {
		loadFixtures(t, cfg.ConnStr, cfg.FixturesDir, tc.Fixtures)
		rec := performRequest(tc, handler)
		assertResponse(t, rec, tc.Response)
	})
}

func loadFixtures(t *testing.T, connStr, fixturesDir string, fixtures []string) {
	for _, fixtureName := range fixtures {
		fixtureName += ".yml"
		fixturePath := filepath.Join(fixturesDir, fixtureName)
		loadFixture(t, connStr, fixturePath)
	}
}

func loadFixture(t *testing.T, connStr, fixturePath string) {
	cfg := &pgfixtures.Config{
		FilePath: fixturePath,
		ConnStr:  connStr,
		Truncate: true,
		ResetSeq: true,
		DryRun:   false,
	}

	err := pgfixtures.Load(context.Background(), cfg)
	if err != nil {
		t.Fatalf("load fixture %s: %v", fixturePath, err)
	}
}

func performRequest(tc TestCase, handler http.Handler) *httptest.ResponseRecorder {
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

	return rec
}

func assertResponse(
	t *testing.T,
	respRecorder *httptest.ResponseRecorder,
	expected ResponseSpec,
) {
	if respRecorder.Result().StatusCode != expected.Status {
		t.Fatalf("unexpected status: got %d, want %d", respRecorder.Result().StatusCode, expected.Status)
	}

	body := respRecorder.Body.String()

	if expected.JSON != "" {
		ja := jsonassert.New(t)
		ja.Assertf(body, expected.JSON)
	}
}
