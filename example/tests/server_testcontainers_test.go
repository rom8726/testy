package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/rom8726/pgfixtures"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rom8726/testy/v2"

	"testyexample"
)

func TestServerWithTestcontainers(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container with testcontainers
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase("db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %v", err)
		}
	}()

	// Get connection string
	// testcontainers returns "postgres://..." but we need "postgresql://..." for compatibility
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}
	// Convert postgres:// to postgresql:// for compatibility with lib/pq
	if len(connStr) > 8 && connStr[:8] == "postgres:" && connStr[8:9] != "q" {
		connStr = "postgresql" + connStr[8:]
	}

	// Apply migration
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationPath := filepath.Join(testDir, "..", "migration.sql")
	if err := applyMigration(ctx, connStr, migrationPath); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	// Setup mocks
	mocks, err := testy.StartMockManager("notification")
	if err != nil {
		t.Fatalf("mock start: %v", err)
	}
	defer mocks.StopAll()

	err = os.Setenv("NOTIFICATION_BASE_URL", mocks.URL("notification"))
	if err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("NOTIFICATION_BASE_URL")

	// Create server instance
	srv := testyexample.NewServer(connStr)

	// Get test directory paths (relative to test file location)
	// Use the same relative paths as in the original test
	casesDir := filepath.Join(testDir, "cases")
	fixturesDir := filepath.Join(testDir, "fixtures")

	// Ensure paths are absolute for reliability
	casesDir, err = filepath.Abs(casesDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for cases: %v", err)
	}
	fixturesDir, err = filepath.Abs(fixturesDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for fixtures: %v", err)
	}

	cfg := testy.Config{
		Handler:     srv.Router,
		DBType:      pgfixtures.PostgreSQL,
		CasesDir:    casesDir,
		FixturesDir: fixturesDir,
		ConnStr:     connStr,
		MockManager: mocks,
		BeforeReq: func() error {
			fmt.Println("before request")
			return nil
		},
		AfterReq: func() error {
			fmt.Println("after request")
			return nil
		},
		JUnitReport: "./junit_testcontainers.xml",
	}

	testy.Run(t, &cfg)
}

// applyMigration applies the migration SQL file to the database
func applyMigration(ctx context.Context, connStr, migrationPath string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Wait for database to be ready
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Read migration file
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if _, err := db.ExecContext(ctx, string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
