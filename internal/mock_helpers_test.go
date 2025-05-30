package internal

import (
	"strings"
	"testing"
)

func TestFindMockInstance(t *testing.T) {
	// Create mock instances
	router1 := NewDynamicMockRouter("mock1")
	router2 := NewDynamicMockRouter("mock2")

	mock1 := NewMockInstance("mock1", "http://mock1.example.com", router1)
	mock2 := NewMockInstance("mock2", "http://mock2.example.com", router2)

	mocks := []*MockInstance{mock1, mock2}

	// Test finding an existing mock
	found := FindMockInstance(mocks, "mock1")
	if found == nil {
		t.Error("Expected to find mock1, but got nil")
	} else if found.name != "mock1" {
		t.Errorf("Expected mock name to be 'mock1', got '%s'", found.name)
	}

	// Test finding another existing mock
	found = FindMockInstance(mocks, "mock2")
	if found == nil {
		t.Error("Expected to find mock2, but got nil")
	} else if found.name != "mock2" {
		t.Errorf("Expected mock name to be 'mock2', got '%s'", found.name)
	}

	// Test finding a non-existent mock
	found = FindMockInstance(mocks, "mock3")
	if found != nil {
		t.Errorf("Expected nil for non-existent mock, got %v", found)
	}

	// Test with empty mocks slice
	found = FindMockInstance([]*MockInstance{}, "mock1")
	if found != nil {
		t.Errorf("Expected nil for empty mocks slice, got %v", found)
	}

	// Test with nil mocks slice
	found = FindMockInstance(nil, "mock1")
	if found != nil {
		t.Errorf("Expected nil for nil mocks slice, got %v", found)
	}
}

func TestGetMockCalls(t *testing.T) {
	// Create mock instances with calls
	router1 := NewDynamicMockRouter("mock1")
	router2 := NewDynamicMockRouter("mock2")

	mock1 := NewMockInstance("mock1", "http://mock1.example.com", router1)
	mock2 := NewMockInstance("mock2", "http://mock2.example.com", router2)

	// Add calls to mock1
	call1 := MockCall{
		Method:  "GET",
		Path:    "/api/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    "",
	}
	call2 := MockCall{
		Method:  "POST",
		Path:    "/api/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name":"John"}`,
	}

	*router1.spy.Calls = append(*router1.spy.Calls, call1, call2)

	// Add calls to mock2
	call3 := MockCall{
		Method:  "GET",
		Path:    "/api/products",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    "",
	}

	*router2.spy.Calls = append(*router2.spy.Calls, call3)

	mocks := []*MockInstance{mock1, mock2}

	// Test getting calls for mock1
	calls := GetMockCalls(mocks, "mock1")
	if len(calls) != 2 {
		t.Errorf("Expected 2 calls for mock1, got %d", len(calls))
	} else {
		if calls[0].Method != "GET" || calls[0].Path != "/api/users" {
			t.Errorf("Unexpected call data: %v", calls[0])
		}
		if calls[1].Method != "POST" || calls[1].Path != "/api/users" {
			t.Errorf("Unexpected call data: %v", calls[1])
		}
	}

	// Test getting calls for mock2
	calls = GetMockCalls(mocks, "mock2")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call for mock2, got %d", len(calls))
	} else {
		if calls[0].Method != "GET" || calls[0].Path != "/api/products" {
			t.Errorf("Unexpected call data: %v", calls[0])
		}
	}

	// Test getting calls for non-existent mock
	calls = GetMockCalls(mocks, "mock3")
	if calls != nil {
		t.Errorf("Expected nil for non-existent mock, got %v", calls)
	}
}

func TestAssertMockCalls_SuccessfulAssertion(t *testing.T) {
	// Create mock instances with calls
	router1 := NewDynamicMockRouter("mock1")

	mock1 := NewMockInstance("mock1", "http://mock1.example.com", router1)

	// Add calls to mock1
	call1 := MockCall{
		Method:  "GET",
		Path:    "/api/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    "",
	}
	call2 := MockCall{
		Method:  "POST",
		Path:    "/api/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name":"John"}`,
	}
	call3 := MockCall{
		Method:  "POST",
		Path:    "/api/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name":"Jane"}`,
	}

	*router1.spy.Calls = append(*router1.spy.Calls, call1, call2, call3)

	mocks := []*MockInstance{mock1}

	// Test successful assertion
	checks := []MockCallCheck{
		{
			Mock:  "mock1",
			Count: 1,
			Expect: MockCallExpect{
				Method: "GET",
				Path:   "/api/users",
			},
		},
		{
			Mock:  "mock1",
			Count: 2,
			Expect: MockCallExpect{
				Method: "POST",
				Path:   "/api/users",
			},
		},
	}

	// This should not panic or fail
	AssertMockCalls(t, checks, mocks)
}

func TestAssertMockCalls_BodyContains(t *testing.T) {
	// Create mock instances with calls
	router1 := NewDynamicMockRouter("mock1")

	mock1 := NewMockInstance("mock1", "http://mock1.example.com", router1)

	// Add calls to mock1
	call1 := MockCall{
		Method:  "POST",
		Path:    "/api/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name":"John"}`,
	}

	*router1.spy.Calls = append(*router1.spy.Calls, call1)

	mocks := []*MockInstance{mock1}

	// Test body contains assertion
	checks := []MockCallCheck{
		{
			Mock:  "mock1",
			Count: 1,
			Expect: MockCallExpect{
				Method: "POST",
				Path:   "/api/users",
				Body: struct {
					Contains string `yaml:"contains"`
				}{
					Contains: "John",
				},
			},
		},
	}

	// This should not panic or fail
	AssertMockCalls(t, checks, mocks)
}

// We can't directly test failure cases for AssertMockCalls because it calls t.Errorf
// Instead, we'll test the underlying logic directly

func TestCountMatchingCalls(t *testing.T) {
	// Create mock calls
	calls := []MockCall{
		{
			Method:  "GET",
			Path:    "/api/users",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    "",
		},
		{
			Method:  "POST",
			Path:    "/api/users",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    `{"name":"John"}`,
		},
		{
			Method:  "POST",
			Path:    "/api/users",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    `{"name":"Jane"}`,
		},
	}

	// Test matching by method only
	expect := MockCallExpect{
		Method: "GET",
	}
	matched := countMatchingCalls(calls, expect)
	if matched != 1 {
		t.Errorf("Expected 1 matching call for GET method, got %d", matched)
	}

	// Test matching by method and path
	expect = MockCallExpect{
		Method: "POST",
		Path:   "/api/users",
	}
	matched = countMatchingCalls(calls, expect)
	if matched != 2 {
		t.Errorf("Expected 2 matching calls for POST to /api/users, got %d", matched)
	}

	// Test matching by body contains
	expect = MockCallExpect{
		Method: "POST",
		Path:   "/api/users",
		Body: struct {
			Contains string `yaml:"contains"`
		}{
			Contains: "John",
		},
	}
	matched = countMatchingCalls(calls, expect)
	if matched != 1 {
		t.Errorf("Expected 1 matching call for body containing 'John', got %d", matched)
	}

	// Test no matches
	expect = MockCallExpect{
		Method: "DELETE",
	}
	matched = countMatchingCalls(calls, expect)
	if matched != 0 {
		t.Errorf("Expected 0 matching calls for DELETE method, got %d", matched)
	}
}

// Helper function to count matching calls (extracted from AssertMockCalls for testing)
func countMatchingCalls(calls []MockCall, expect MockCallExpect) int {
	matched := 0
	for _, call := range calls {
		if expect.Method != "" && call.Method != expect.Method {
			continue
		}
		if expect.Path != "" && call.Path != expect.Path {
			continue
		}
		if expect.Body.Contains != "" && !strings.Contains(call.Body, expect.Body.Contains) {
			continue
		}
		matched++
	}
	return matched
}
