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
func LoadFixturesFromList(t *testing.T, dbType pgfixtures.DatabaseType, connStr, fixturesDir string, fixtures []string) {
	t.Helper()

	for _, fixtureName := range fixtures {
		fixtureName += ".yml"
		fixturePath := filepath.Join(fixturesDir, fixtureName)
		LoadFixtureFile(t, dbType, connStr, fixturePath)
	}
}

// LoadFixtureFile loads a single fixture file
func LoadFixtureFile(t *testing.T, dbType pgfixtures.DatabaseType, connStr, fixturePath string) {
	t.Helper()

	cfg := &pgfixtures.Config{
		FilePath:     fixturePath,
		ConnStr:      connStr,
		DatabaseType: dbType,
		Truncate:     true,
		ResetSeq:     true,
		DryRun:       false,
	}

	err := pgfixtures.Load(t.Context(), cfg)
	if err != nil {
		dbErr := NewError(ErrDatabase, "LoadFixtureFile", "failed to load fixture").
			WithContext("fixture", fixturePath).
			WithContext("error", err.Error())
		t.Fatalf("%v", dbErr)
	}
}

// ExecuteDBChecks executes all database checks for a step
func ExecuteDBChecks(t *testing.T, connStr string, step Step, ctxMap map[string]any) {
	t.Helper()
	const op = "ExecuteDBChecks"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		dbErr := NewError(ErrDatabase, op, "failed to open database connection").
			WithContext("step", step.Name).
			WithContext("error", err.Error())
		t.Fatalf("%v", dbErr)
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
	const op = "ExecuteDBCheck"

	rows, err := db.QueryContext(t.Context(), check.Query)
	if err != nil {
		dbErr := NewError(ErrDatabase, op, "failed to execute database query").
			WithContext("query", check.Query).
			WithContext("error", err.Error())
		t.Fatalf("%v", dbErr)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		dbErr := NewError(ErrDatabase, op, "failed to get column information").
			WithContext("error", err.Error())
		t.Fatalf("%v", dbErr)
	}

	results := make([]map[string]any, 0)

	for rows.Next() {
		row := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range ptrs {
			ptrs[i] = &row[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			dbErr := NewError(ErrDatabase, op, "failed to scan row data").
				WithContext("error", err.Error())
			t.Fatalf("%v", dbErr)
		}

		m := make(map[string]any)
		for i, col := range cols {
			m[col] = row[i]
		}

		results = append(results, m)
	}

	actual, err := json.Marshal(results)
	if err != nil {
		dbErr := NewError(ErrInternal, op, "failed to marshal query results").
			WithContext("error", err.Error())
		t.Fatalf("%v", dbErr)
	}

	var expectedJSON string
	switch v := check.Result.(type) {
	case string:
		expectedJSON = v
	default:
		buf, err := json.Marshal(v)
		if err != nil {
			dbErr := NewError(ErrInternal, op, "failed to marshal expected results").
				WithContext("error", err.Error())
			t.Fatalf("%v", dbErr)
		}

		expectedJSON = string(buf)
	}

	ja := jsonassert.New(t)
	ja.Assert(string(actual), expectedJSON)
}
