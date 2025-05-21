package internal

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kinbiko/jsonassert"
	_ "github.com/lib/pq"
	"github.com/rom8726/pgfixtures"
)

func RunSingle(t *testing.T, handler http.Handler, tc TestCase, cfg *Config) {
	t.Helper()

	t.Run(tc.Name, func(t *testing.T) {
		loadFixtures(t, cfg.ConnStr, cfg.FixturesDir, tc.Fixtures)

		if cfg.BeforeReq != nil {
			if err := cfg.BeforeReq(); err != nil {
				t.Fatalf("beforeReq failed: %v", err)
			}
		}

		rec := performRequest(t, tc, handler)

		if cfg.AfterReq != nil {
			if err := cfg.AfterReq(); err != nil {
				t.Fatalf("afterReq failed: %v", err)
			}
		}

		assertResponse(t, rec, tc.Response)

		performDBChecks(t, cfg.ConnStr, tc)
	})
}

func loadFixtures(t *testing.T, connStr, fixturesDir string, fixtures []string) {
	t.Helper()

	for _, fixtureName := range fixtures {
		fixtureName += ".yml"
		fixturePath := filepath.Join(fixturesDir, fixtureName)
		loadFixture(t, connStr, fixturePath)
	}
}

func loadFixture(t *testing.T, connStr, fixturePath string) {
	t.Helper()

	cfg := &pgfixtures.Config{
		FilePath: fixturePath,
		ConnStr:  connStr,
		Truncate: true,
		ResetSeq: true,
		DryRun:   false,
	}

	err := pgfixtures.Load(t.Context(), cfg)
	if err != nil {
		t.Fatalf("load fixture %s: %v", fixturePath, err)
	}
}

func performRequest(t *testing.T, tc TestCase, handler http.Handler) *httptest.ResponseRecorder {
	t.Helper()

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

func performDBChecks(t *testing.T, connStr string, tc TestCase) {
	t.Helper()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("cannot open db: %v", err)
	}
	defer db.Close()

	for i, check := range tc.DBChecks {
		performDBCheck(t, db, i, check)
	}
}

func performDBCheck(t *testing.T, db *sql.DB, idx int, check DBCheck) {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), check.Query)
	if err != nil {
		t.Fatalf("dbCheck failed for query %q: %v", check.Query, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}

	results := make([]map[string]any, 0)

	for rows.Next() {
		row := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range ptrs {
			ptrs[i] = &row[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan error: %v", err)
		}

		m := make(map[string]any)
		for i, col := range cols {
			m[col] = row[i]
		}

		results = append(results, m)
	}

	actual, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("cannot marshal dbCheck result: %v", err)
	}

	var expectedJSON string
	switch v := check.Result.(type) {
	case string:
		expectedJSON = v
	default:
		buf, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("cannot marshal expected dbCheck result: %v", err)
		}

		expectedJSON = string(buf)
	}

	ja := jsonassert.New(t)
	ja.Assert(string(actual), expectedJSON)
}
