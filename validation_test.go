package testy

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rom8726/pgfixtures"
)

func TestConfig_Validate(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "testy-validation-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	casesDir := filepath.Join(tempDir, "cases")
	fixturesDir := filepath.Join(tempDir, "fixtures")

	if err := os.MkdirAll(casesDir, 0755); err != nil {
		t.Fatalf("Failed to create cases dir: %v", err)
	}
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatalf("Failed to create fixtures dir: %v", err)
	}

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: &Config{
				Handler:     handler,
				CasesDir:    casesDir,
				FixturesDir: fixturesDir,
				ConnStr:     "postgresql://user:password@localhost:5432/db",
				DBType:      pgfixtures.PostgreSQL,
			},
			wantErr: false,
		},
		{
			name: "missing handler",
			config: &Config{
				CasesDir:    casesDir,
				FixturesDir: fixturesDir,
				ConnStr:     "postgresql://user:password@localhost:5432/db",
				DBType:      pgfixtures.PostgreSQL,
			},
			wantErr:     true,
			errContains: "Handler",
		},
		{
			name: "missing cases dir",
			config: &Config{
				Handler:     handler,
				FixturesDir: fixturesDir,
				ConnStr:     "postgresql://user:password@localhost:5432/db",
				DBType:      pgfixtures.PostgreSQL,
			},
			wantErr:     true,
			errContains: "CasesDir",
		},
		{
			name: "non-existent cases dir",
			config: &Config{
				Handler:     handler,
				CasesDir:    "/non/existent/path",
				FixturesDir: fixturesDir,
				ConnStr:     "postgresql://user:password@localhost:5432/db",
				DBType:      pgfixtures.PostgreSQL,
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name: "non-existent fixtures dir",
			config: &Config{
				Handler:     handler,
				CasesDir:    casesDir,
				FixturesDir: "/non/existent/fixtures",
				ConnStr:     "postgresql://user:password@localhost:5432/db",
				DBType:      pgfixtures.PostgreSQL,
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name: "fixtures dir without connection string",
			config: &Config{
				Handler:     handler,
				CasesDir:    casesDir,
				FixturesDir: fixturesDir,
			},
			wantErr:     true,
			errContains: "connection string is required",
		},
		{
			name: "connection string without db type",
			config: &Config{
				Handler:     handler,
				CasesDir:    casesDir,
				FixturesDir: fixturesDir,
				ConnStr:     "postgresql://user:password@localhost:5432/db",
			},
			wantErr:     true,
			errContains: "DBType must be specified",
		},
		{
			name: "valid config without fixtures",
			config: &Config{
				Handler:  handler,
				CasesDir: casesDir,
			},
			wantErr: false,
		},
		{
			name: "junit report with auto-create directory",
			config: &Config{
				Handler:     handler,
				CasesDir:    casesDir,
				JUnitReport: filepath.Join(tempDir, "reports", "junit.xml"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "TestField",
		Message: "test message",
	}

	expected := "validation error: TestField - test message"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}
