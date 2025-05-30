package internal

import (
	"strings"
	"testing"
)

// AssertMockCalls checks if the mock calls match the expected calls
func AssertMockCalls(t *testing.T, checks []MockCallCheck, mocks []*MockInstance) {
	t.Helper()

	for _, check := range checks {
		calls := GetMockCalls(mocks, check.Mock)
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

// GetMockCalls returns all calls made to a mock
func GetMockCalls(mocks []*MockInstance, name string) []MockCall {
	for _, inst := range mocks {
		if inst.name == name {
			return *inst.router.spy.Calls
		}
	}

	return nil
}

// FindMockInstance returns a mock instance by name
func FindMockInstance(mocks []*MockInstance, name string) *MockInstance {
	for _, inst := range mocks {
		if inst.name == name {
			return inst
		}
	}

	return nil
}
