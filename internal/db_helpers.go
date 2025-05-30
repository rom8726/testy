package internal

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/kinbiko/jsonassert"
	_ "github.com/lib/pq"
	"github.com/rom8726/pgfixtures"
)

// LoadFixturesFromList loads all fixtures specified in the test case
func LoadFixturesFromList(t *testing.T, connStr, fixturesDir string, fixtures []string) {
	t.Helper()

	for _, fixtureName := range fixtures {
		fixtureName += ".yml"
		fixturePath := filepath.Join(fixturesDir, fixtureName)
		LoadFixtureFile(t, connStr, fixturePath)
	}
}

// LoadFixtureFile loads a single fixture file
func LoadFixtureFile(t *testing.T, connStr, fixturePath string) {
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

// ExecuteDBChecks executes all database checks for a step
func ExecuteDBChecks(t *testing.T, connStr string, step Step, ctxMap map[string]any) {
	t.Helper()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("cannot open db: %v", err)
	}
	defer db.Close()

	for _, check := range step.DBChecks {
		check = renderDBCheck(check, ctxMap)
		ExecuteDBCheck(t, db, check)
	}
}

// ExecuteDBCheck executes a single database check
func ExecuteDBCheck(t *testing.T, db *sql.DB, check DBCheck) {
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
