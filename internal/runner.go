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
	const op = "RunSingle"

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
				mockErr := NewError(ErrMock, op, "mock not found").
					WithContext("mock", name)
				t.Fatalf("%+v", mockErr)
			}

			for _, route := range def.Routes {
				inst.router.AddRoute(route)
			}

			ctxMap[name+".baseURL"] = inst.url
			ctxMap[name+".calls"] = inst.router.spy.Calls
		}

		// Load fixtures
		LoadFixturesFromList(t, cfg.DBType, cfg.ConnStr, cfg.FixturesDir, tc.Fixtures)

		for _, step := range tc.Steps {
			step.Name = strings.ReplaceAll(step.Name, " ", "_")
			if step.Response.JSON != "" && step.Response.Headers == nil {
				step.Response.Headers = map[string]string{
					"Content-Type": "application/json; charset=utf-8",
				}
			}

			performStep(t, handler, step, cfg, ctxMap)
		}

		// Assert mock calls
		AssertMockCalls(t, tc.MockCalls, cfg.Mocks)
	})

	return res
}

// performStep executes a single step in a test case
func performStep(t *testing.T, handler http.Handler, step Step, cfg *Config, ctxMap map[string]any) {
	t.Helper()
	const op = "performStep"

	if cfg.BeforeReq != nil {
		if err := cfg.BeforeReq(); err != nil {
			hookErr := NewError(ErrInternal, op, "beforeReq hook failed").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", hookErr)
		}
	}

	// Execute request using HTTP client
	rec := ExecuteRequest(t, step, handler, ctxMap)

	if cfg.AfterReq != nil {
		if err := cfg.AfterReq(); err != nil {
			hookErr := NewError(ErrInternal, op, "afterReq hook failed").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", hookErr)
		}
	}

	// Extract JSON fields from the response body
	if rec != nil && rec.Body != nil {
		respBody := rec.Body.Bytes()
		if len(respBody) > 0 {
			if step.Response.Headers != nil && strings.HasPrefix(step.Response.Headers["Content-Type"], "application/json") {
				var jsonData any
				if err := json.Unmarshal(respBody, &jsonData); err != nil {
					jsonErr := NewError(ErrHTTP, op, "failed to parse response JSON").
						WithContext("step", step.Name).
						WithContext("error", err.Error())
					t.Fatalf("%+v", jsonErr)
				}
				extractJSONFields(step.Name+".response", jsonData, ctxMap)
			}
		}
	}

	// Assert response using HTTP client
	AssertResponse(t, rec, step.Response)

	// Execute DB checks using database client
	ExecuteDBChecks(t, cfg.ConnStr, step, ctxMap)
}
