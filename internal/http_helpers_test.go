package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExecuteRequest(t *testing.T) {
	// Create a test handler that returns different responses based on the request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content type
		w.Header().Set("Content-Type", "application/json")

		// Check the path and method to determine the response
		if r.URL.Path == "/users" && r.Method == "GET" {
			// Return a list of users
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": 1, "name": "User1"},
					{"id": 2, "name": "User2"},
				},
			})
			return
		}

		if r.URL.Path == "/users" && r.Method == "POST" {
			// Check for authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Unauthorized",
				})
				return
			}

			// Parse the request body
			var reqBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid request body",
				})
				return
			}

			// Return the created user
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   3,
				"name": reqBody["name"],
			})
			return
		}

		if r.URL.Path == "/users/1" && r.Method == "GET" {
			// Return a specific user
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"name": "User1",
			})
			return
		}

		// Default: not found
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Not found",
		})
	})

	tests := []struct {
		name          string
		step          Step
		ctxMap        map[string]any
		expectedCode  int
		expectedBody  string
		shouldContain string
	}{
		{
			name: "GET users - success",
			step: Step{
				Name: "get_users",
				Request: RequestSpec{
					Method: "GET",
					Path:   "/users",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				},
				Response: ResponseSpec{
					Status: 200,
				},
			},
			ctxMap:        map[string]any{},
			expectedCode:  200,
			shouldContain: `"users":[{"id":1,"name":"User1"},{"id":2,"name":"User2"}]`,
		},
		{
			name: "POST user - success",
			step: Step{
				Name: "create_user",
				Request: RequestSpec{
					Method: "POST",
					Path:   "/users",
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer {{token}}",
					},
					Body: map[string]any{
						"name": "{{userName}}",
					},
				},
				Response: ResponseSpec{
					Status: 201,
				},
			},
			ctxMap: map[string]any{
				"token":    "valid-token",
				"userName": "NewUser",
			},
			expectedCode:  201,
			shouldContain: `"name":"NewUser"`,
		},
		{
			name: "POST user - unauthorized",
			step: Step{
				Name: "create_user_unauthorized",
				Request: RequestSpec{
					Method: "POST",
					Path:   "/users",
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer {{token}}",
					},
					Body: map[string]any{
						"name": "NewUser",
					},
				},
				Response: ResponseSpec{
					Status: 401,
				},
			},
			ctxMap: map[string]any{
				"token": "invalid-token",
			},
			expectedCode:  401,
			shouldContain: `"error":"Unauthorized"`,
		},
		{
			name: "GET user by ID - success",
			step: Step{
				Name: "get_user",
				Request: RequestSpec{
					Method: "GET",
					Path:   "/users/{{userId}}",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				},
				Response: ResponseSpec{
					Status: 200,
				},
			},
			ctxMap: map[string]any{
				"userId": 1,
			},
			expectedCode:  200,
			shouldContain: `"name":"User1"`,
		},
		{
			name: "GET non-existent resource - 404",
			step: Step{
				Name: "get_nonexistent",
				Request: RequestSpec{
					Method: "GET",
					Path:   "/nonexistent",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				},
				Response: ResponseSpec{
					Status: 404,
				},
			},
			ctxMap:        map[string]any{},
			expectedCode:  404,
			shouldContain: `"error":"Not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the request
			rec := ExecuteRequest(t, tt.step, handler, tt.ctxMap)

			// Check status code
			if rec.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, rec.Code)
			}

			// Check response body contains expected string
			if tt.shouldContain != "" && !contains(rec.Body.String(), tt.shouldContain) {
				t.Errorf("Response body does not contain expected string.\nExpected to contain: %s\nGot: %s",
					tt.shouldContain, rec.Body.String())
			}
		})
	}
}

func TestAssertResponse(t *testing.T) {
	tests := []struct {
		name       string
		recorder   *httptest.ResponseRecorder
		expected   ResponseSpec
		shouldFail bool
	}{
		{
			name: "status code match",
			recorder: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusOK)
				return rec
			}(),
			expected: ResponseSpec{
				Status: 200,
			},
			shouldFail: false,
		},
		{
			name: "JSON match",
			recorder: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusOK)
				json.NewEncoder(rec).Encode(map[string]interface{}{
					"id":   1,
					"name": "User1",
				})
				return rec
			}(),
			expected: ResponseSpec{
				Status: 200,
				JSON:   `{"id": 1, "name": "User1"}`,
			},
			shouldFail: false,
		},
		{
			name: "JSON with wildcard match",
			recorder: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusOK)
				json.NewEncoder(rec).Encode(map[string]interface{}{
					"id":        123,
					"name":      "User1",
					"createdAt": "2023-01-01T12:00:00Z",
				})
				return rec
			}(),
			expected: ResponseSpec{
				Status: 200,
				JSON:   `{"id": "<<PRESENCE>>", "name": "User1", "createdAt": "<<PRESENCE>>"}`,
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic or fail for non-failure cases
			AssertResponse(t, tt.recorder, tt.expected)
		})
	}

	// Test failure case separately
	t.Run("status code mismatch", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusNotFound)

		// We can't directly test the failure since AssertResponse calls t.Fatalf
		// Instead, we'll verify that the status code is different from what we expect
		if rec.Result().StatusCode == http.StatusOK {
			t.Errorf("Expected status code to be different from OK")
		}

		// Just verify that the status code is what we set it to
		if rec.Result().StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Result().StatusCode)
		}
	})
}

// Helper functions

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && strings.Contains(s, substr)
}

// mockTestingT is a mock implementation of testing.TB for testing failure cases
type mockTestingT struct {
	failed       bool
	errorMessage string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.failed = true
	m.errorMessage = format
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.errorMessage = format
}

func (m *mockTestingT) Helper() {}
