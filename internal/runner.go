package internal

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	_ "github.com/lib/pq"
	"github.com/rom8726/pgfixtures"
)

func RunSingle(t *testing.T, handler http.Handler, tc TestCase, cfg *Config) {
	t.Helper()

	t.Run(tc.Name, func(t *testing.T) {
		ctxMap := initCtxMap()
		for _, inst := range cfg.Mocks {
			ctxMap[inst.name+".baseURL"] = inst.url
			ctxMap[inst.name+".calls"] = inst.spy.Calls
		}

		loadFixtures(t, cfg.ConnStr, cfg.FixturesDir, tc.Fixtures)

		for _, step := range tc.Steps {
			step.Name = strings.ReplaceAll(step.Name, " ", "_")

			performStep(t, handler, step, cfg, ctxMap)
		}

		assertMockCalls(t, tc.MockCalls, cfg.Mocks)
	})
}

func performStep(t *testing.T, handler http.Handler, step Step, cfg *Config, ctxMap map[string]any) {
	t.Helper()

	if cfg.BeforeReq != nil {
		if err := cfg.BeforeReq(); err != nil {
			t.Fatalf("beforeReq failed for step %q: %v", step.Name, err)
		}
	}

	rec := performRequest(t, step, handler, ctxMap)

	if cfg.AfterReq != nil {
		if err := cfg.AfterReq(); err != nil {
			t.Fatalf("afterReq failed for step %q: %v", step.Name, err)
		}
	}

	if respBody := rec.Body.Bytes(); len(respBody) > 0 {
		var jsonData any
		if err := json.Unmarshal(respBody, &jsonData); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		extractJSONFields(step.Name+".response", jsonData, ctxMap)
	}

	assertResponse(t, rec, step.Response)

	performDBChecks(t, cfg.ConnStr, step, ctxMap)
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

func performRequest(t *testing.T, step Step, handler http.Handler, ctxMap map[string]any) *httptest.ResponseRecorder {
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

func performDBChecks(t *testing.T, connStr string, step Step, ctxMap map[string]any) {
	t.Helper()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("cannot open db: %v", err)
	}
	defer db.Close()

	for _, check := range step.DBChecks {
		check = renderDBCheck(check, ctxMap)
		performDBCheck(t, db, check)
	}
}

func performDBCheck(t *testing.T, db *sql.DB, check DBCheck) {
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

func assertMockCalls(t *testing.T, checks []MockCallCheck, mocks []*MockInstance) {
	t.Helper()

	for _, check := range checks {
		calls := mockCalls(mocks, check.Mock)
		if calls == nil {
			t.Errorf("mock %q not found or has no calls", check.Mock)

			continue
		}

		matched := 0
		for _, call := range calls {
			if check.Expect.Method != "" && call.Method != check.Expect.Method {
				continue
			}
			if check.Expect.Path != "" && call.Path != check.Expect.Path {
				continue
			}
			if check.Expect.Body.Contains != "" && !strings.Contains(call.Body, check.Expect.Body.Contains) {
				continue
			}
			matched++
		}

		if matched != check.Count {
			t.Errorf("mock %q expected %d matching calls, got %d", check.Mock, check.Count, matched)
		}
	}
}

func mockCalls(mocks []*MockInstance, name string) []MockCall {
	for _, inst := range mocks {
		if inst.name == name {
			return *inst.spy.Calls
		}
	}

	return nil
}
