package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTestCases(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "testy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test YAML files
	testCases := []struct {
		filename string
		content  string
		valid    bool
	}{
		{
			filename: "valid1.yml",
			content: `
- name: Test Case 1
  steps:
    - name: step1
      request:
        method: GET
        path: /api/users
      response:
        status: 200
`,
			valid: true,
		},
		{
			filename: "valid2.yml",
			content: `
- name: Test Case 2
  fixtures:
    - users
  steps:
    - name: step1
      request:
        method: POST
        path: /api/users
        body:
          name: John
      response:
        status: 201
      dbChecks:
        - query: SELECT COUNT(*) FROM users
          result: '[{"count": 1}]'
`,
			valid: true,
		},
		{
			filename: "not-yaml.txt",
			content:  "This is not a YAML file",
			valid:    false, // Should be ignored because it's not a .yml file
		},
	}

	// Write test files to the temporary directory
	for _, tc := range testCases {
		filePath := filepath.Join(tempDir, tc.filename)
		err := os.WriteFile(filePath, []byte(tc.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", tc.filename, err)
		}
	}

	// Test loading valid files
	cases, err := LoadTestCases(tempDir)
	if err != nil {
		t.Fatalf("LoadTestCases returned error: %v", err)
	}

	// Verify the number of test cases
	// We expect 2 test cases from valid1.yml and valid2.yml
	expectedCount := 2
	if len(cases) != expectedCount {
		t.Errorf("Expected %d test cases, got %d", expectedCount, len(cases))
	}

	// Verify the content of the test cases
	if len(cases) > 0 {
		if cases[0].Name != "Test Case 1" {
			t.Errorf("Expected name 'Test Case 1', got '%s'", cases[0].Name)
		}
		if len(cases[0].Steps) != 1 {
			t.Errorf("Expected 1 step, got %d", len(cases[0].Steps))
		}
		if cases[0].Steps[0].Request.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", cases[0].Steps[0].Request.Method)
		}
	}

	if len(cases) > 1 {
		if cases[1].Name != "Test Case 2" {
			t.Errorf("Expected name 'Test Case 2', got '%s'", cases[1].Name)
		}
		if len(cases[1].Fixtures) != 1 || cases[1].Fixtures[0] != "users" {
			t.Errorf("Expected fixtures ['users'], got %v", cases[1].Fixtures)
		}
		if len(cases[1].Steps[0].DBChecks) != 1 {
			t.Errorf("Expected 1 DB check, got %d", len(cases[1].Steps[0].DBChecks))
		}
	}

	t.Run("loading invalid YAML", func(t *testing.T) {
		// Test loading invalid YAML
		invalidDir := filepath.Join(tempDir, "invalid")
		err = os.Mkdir(invalidDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create invalid dir: %v", err)
		}

		invalidFilePath := filepath.Join(invalidDir, "invalid.yml")
		err = os.WriteFile(invalidFilePath, []byte("invalid yaml: {"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid test file: %v", err)
		}

		_, err = LoadTestCases(invalidDir)
		if err == nil {
			t.Error("Expected error when loading invalid YAML, got nil")
		}
	})

	t.Run("loading non-existent directory", func(t *testing.T) {
		// Test loading from non-existent directory
		_, err = LoadTestCases("/non/existent/directory")
		if err == nil {
			t.Error("Expected error when loading from non-existent directory, got nil")
		}
	})
}
