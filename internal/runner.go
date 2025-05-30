package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestCaseResult represents the result of running a test case
type TestCaseResult struct {
	Name     string
	Duration time.Duration
	ErrMsg   string
}

// RunSingle runs a single test case
func RunSingle(t *testing.T, handler http.Handler, tc TestCase, cfg *Config) TestCaseResult {
	t.Helper()

	start := time.Now()
	res := TestCaseResult{Name: tc.Name}

	defer func() {
		res.Duration = time.Since(start)
		if r := recover(); r != nil {
			res.ErrMsg = fmt.Sprint(r)
			panic(r)
		}
	}()

	t.Run(tc.Name, func(t *testing.T) {
		ctxMap := initCtxMap()
		for name, def := range tc.Mocks {
			inst := FindMockInstance(cfg.Mocks, name)
			if inst == nil {
				t.Fatalf("mock %q not found", name)
			}

			for _, route := range def.Routes {
				inst.router.AddRoute(route)
			}

			ctxMap[name+".baseURL"] = inst.url
			ctxMap[name+".calls"] = inst.router.spy.Calls
		}

		LoadFixturesFromList(t, cfg.ConnStr, cfg.FixturesDir, tc.Fixtures)

		for _, step := range tc.Steps {
			step.Name = strings.ReplaceAll(step.Name, " ", "_")

			performStep(t, handler, step, cfg, ctxMap)
		}

		AssertMockCalls(t, tc.MockCalls, cfg.Mocks)
	})

	return res
}

// performStep executes a single step in a test case
func performStep(t *testing.T, handler http.Handler, step Step, cfg *Config, ctxMap map[string]any) {
	t.Helper()

	if cfg.BeforeReq != nil {
		if err := cfg.BeforeReq(); err != nil {
			t.Fatalf("beforeReq failed for step %q: %v", step.Name, err)
		}
	}

	rec := ExecuteRequest(t, step, handler, ctxMap)

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

	AssertResponse(t, rec, step.Response)

	ExecuteDBChecks(t, cfg.ConnStr, step, ctxMap)
}
