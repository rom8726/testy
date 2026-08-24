package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
)

// TestCaseResult represents the result of running a test case
type TestCaseResult struct {
	Name     string
	Duration time.Duration
	ErrMsg   string
}

// RunSingle runs a single test case with all features integrated
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
		// Initialize context with environment variables
		ctxMap := initCtxMap()

		// Add test case variables
		if tc.Variables != nil {
			for k, v := range tc.Variables {
				ctxMap[k] = v
			}
		}

		// Setup mocks
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

		// Setup hook executor
		var db *sql.DB
		var hookExecutor *HookExecutor
		var testServer *httptest.Server

		// Create test server for HTTP hooks if handler is provided
		if handler != nil {
			testServer = httptest.NewServer(handler)
			defer testServer.Close()
		}

		if cfg.ConnStr != "" {
			var err error
			db, err = sql.Open("postgres", cfg.ConnStr)
			if err == nil {
				defer func() { _ = db.Close() }()
			}
		}

		// Create hook executor with test server URL if available
		baseURL := ""
		if testServer != nil {
			baseURL = testServer.URL
		}
		hookExecutor = NewHookExecutor(db, baseURL)

		// Execute setup hooks
		if len(tc.Setup) > 0 && hookExecutor != nil {
			if err := ExecuteSetup(hookExecutor, tc.Setup, ctxMap); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
		}

		// Setup deferred teardown
		if len(tc.Teardown) > 0 && hookExecutor != nil {
			defer func() {
				if err := ExecuteTeardown(hookExecutor, tc.Teardown, ctxMap); err != nil {
					t.Logf("Warning: teardown failed: %v", err)
				}
			}()
		}

		// Load fixtures
		LoadFixturesFromList(t, cfg.DBType, cfg.ConnStr, cfg.FixturesDir, tc.Fixtures)

		// Execute steps
		for _, step := range tc.Steps {
			step.Name = strings.ReplaceAll(step.Name, " ", "_")

			// Check if a step should execute (conditional)
			if step.When != "" {
				shouldExecute, err := ShouldExecuteStep(step, ctxMap)
				if err != nil {
					t.Fatalf("Failed to evaluate condition for step %s: %v", step.Name, err)
				}
				if !shouldExecute {
					t.Logf("Skipping step %s (condition not met)", step.Name)
					continue
				}
			}

			// Handle loops
			if step.Loop != nil {
				contexts, err := ExpandLoop(step.Loop, ctxMap)
				if err != nil {
					t.Fatalf("Failed to expand loop for step %s: %v", step.Name, err)
				}

				for i, loopCtx := range contexts {
					loopStepName := fmt.Sprintf("%s[%d]", step.Name, i)
					loopStep := step
					loopStep.Name = loopStepName
					performStep(t, handler, loopStep, cfg, loopCtx, db)
				}
			} else {
				performStep(t, handler, step, cfg, ctxMap, db)
			}
		}

		// Assert mock calls
		AssertMockCalls(t, tc.MockCalls, cfg.Mocks)
	})

	return res
}

// performStep executes a single step with all features
func performStep(t *testing.T, handler http.Handler, step Step, cfg *Config, ctxMap map[string]any, db *sql.DB) {
	t.Helper()
	const op = "performStep"

	// Expand faker placeholders in context (modify in place to preserve extracted fields)
	expanded := expandFakerInContext(ctxMap)
	for k, v := range expanded {
		ctxMap[k] = v
	}

	// BeforeReq hook
	if cfg.BeforeReq != nil {
		if err := cfg.BeforeReq(); err != nil {
			hookErr := NewError(ErrInternal, op, "beforeReq hook failed").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", hookErr)
		}
	}

	// Determine if this is HTTP or gRPC step
	isHTTP := step.Request != nil
	isGRPC := step.GRPCRequest != nil

	if !isHTTP && !isGRPC {
		stepErr := NewError(ErrInternal, op, "step must have either request or grpcRequest").
			WithContext("step", step.Name)
		t.Fatalf("%+v", stepErr)
	}

	if isHTTP && isGRPC {
		stepErr := NewError(ErrInternal, op, "step cannot have both request and grpcRequest").
			WithContext("step", step.Name)
		t.Fatalf("%+v", stepErr)
	}

	var requestDuration time.Duration
	var statusCode uint64

	if isHTTP {
		// HTTP request execution
		statusCode = performHTTPStep(t, handler, step, cfg, ctxMap, db, &requestDuration)
	} else {
		// gRPC request execution
		statusCode = performGRPCStep(t, step, cfg, ctxMap, db, &requestDuration)
	}

	// AfterReq hook
	if cfg.AfterReq != nil {
		if err := cfg.AfterReq(); err != nil {
			hookErr := NewError(ErrInternal, op, "afterReq hook failed").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", hookErr)
		}
	}

	// Performance assertions
	if step.Performance != nil {
		AssertPerformance(t, &step, requestDuration, statusCode)
	}
}

func performHTTPStep(
	t *testing.T,
	handler http.Handler,
	step Step,
	_ *Config,
	ctxMap map[string]any,
	db *sql.DB,
	requestDuration *time.Duration,
) uint64 {
	t.Helper()

	var rec *httptest.ResponseRecorder

	// Execute request with retry if configured
	if step.Retry != nil {
		parsedRetry, err := ParseRetryConfig(*step.Retry)
		if err != nil {
			t.Fatalf("Failed to parse retry config: %v", err)
		}

		expectedStatus := step.Response.Status
		result := ExecuteWithRetry(parsedRetry, func() (int, error) {
			startTime := time.Now()
			rec = ExecuteRequest(t, step, handler, ctxMap)
			*requestDuration = time.Since(startTime)

			if rec.Code == expectedStatus {
				return rec.Code, nil
			}

			if rec.Code >= 400 {
				shouldRetry := false
				if len(parsedRetry.RetryOn) > 0 {
					for _, retryCode := range parsedRetry.RetryOn {
						if rec.Code == retryCode {
							shouldRetry = true
							break
						}
					}
				}
				if parsedRetry.RetryOnError {
					shouldRetry = true
				}

				if shouldRetry {
					return rec.Code, fmt.Errorf("HTTP error: %d (expected: %d)", rec.Code, expectedStatus)
				}
				return rec.Code, nil
			}

			return rec.Code, nil
		})

		if !result.Success {
			t.Logf("Request failed after %d attempts", result.Attempts)
		} else if result.Attempts > 1 {
			t.Logf("Request succeeded after %d attempts", result.Attempts)
		}
	} else {
		// Single request execution
		startTime := time.Now()
		rec = ExecuteRequest(t, step, handler, ctxMap)
		*requestDuration = time.Since(startTime)
	}

	// Extract response fields to context
	if rec.Code >= 200 && rec.Code < 300 {
		contentType := rec.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			var jsonData any
			if err := json.Unmarshal(rec.Body.Bytes(), &jsonData); err == nil {
				extractJSONFields(step.Name+".response", jsonData, ctxMap)
			}
		}
	}

	diagnosticStep := renderStepForDiagnostics(step, ctxMap)

	// Assert response
	AssertStepResponse(t, diagnosticStep, rec, *step.Response)

	// JSON Schema validation
	if step.Response.JSONSchema != nil || step.Response.Schema != "" {
		var schema JSONSchema
		var err error

		if step.Response.Schema != "" {
			schema, err = LoadJSONSchemaFromFile(step.Response.Schema)
			if err != nil {
				t.Fatalf("Failed to load schema: %v", err)
			}
		} else {
			schema = *step.Response.JSONSchema
		}

		var jsonData any
		if err := json.Unmarshal(rec.Body.Bytes(), &jsonData); err != nil {
			t.Fatalf("Failed to parse JSON for schema validation: %v", err)
		}

		errors := ValidateJSONSchema(jsonData, schema, "root")
		if len(errors) > 0 {
			t.Errorf("JSON Schema validation failed:")
			for _, err := range errors {
				t.Errorf("  - %s", err.Error())
			}
		}
	}

	// Enhanced assertions
	if len(step.Response.Assertions) > 0 {
		var responseData map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &responseData); err != nil {
			t.Logf("Warning: cannot parse response for assertions: %v", err)
		} else {
			for _, assertion := range step.Response.Assertions {
				renderedAssertion := assertion
				renderedAssertion.Value = RenderAny(assertion.Value, ctxMap)
				if err := AssertResponseV2(renderedAssertion, responseData); err != nil {
					t.Errorf("Assertion failed: %v\n%s", err, formatHTTPDiagnostic(diagnosticStep, rec))
				}
			}
		}
	}

	// Execute DB checks
	if len(step.DBChecks) > 0 && db != nil {
		for _, check := range step.DBChecks {
			check = renderDBCheck(check, ctxMap)
			ExecuteStepDBCheck(t, step.Name, db, check)
		}
	}

	return uint64(rec.Code)
}

// New function for gRPC step execution
func performGRPCStep(
	t *testing.T,
	step Step,
	cfg *Config,
	ctxMap map[string]any,
	db *sql.DB,
	requestDuration *time.Duration,
) uint64 {
	t.Helper()
	const op = "performGRPCStep"

	if cfg.GRPCAddr == "" {
		grpcErr := NewError(ErrGRPC, op, "gRPC address not configured").
			WithContext("step", step.Name)
		t.Fatalf("%+v", grpcErr)
	}

	var response *GRPCResponse

	// Execute gRPC request
	startTime := time.Now()
	response = ExecuteGRPCRequest(t, step, cfg.GRPCAddr, ctxMap)
	*requestDuration = time.Since(startTime)

	// Extract response fields to context
	if response.Code == codes.OK && response.Message != nil {
		// Convert proto message to JSON for context extraction
		jsonBytes, err := protojson.Marshal(response.Message)
		if err == nil {
			var jsonData any
			if err := json.Unmarshal(jsonBytes, &jsonData); err == nil {
				extractJSONFields(step.Name+".response", jsonData, ctxMap)
			}
		}
	}

	// Assert gRPC response
	diagnosticStep := renderStepForDiagnostics(step, ctxMap)
	AssertStepGRPCResponse(t, diagnosticStep, response, *step.GRPCResponse, ctxMap)

	// Enhanced assertions for gRPC
	if len(step.GRPCResponse.Assertions) > 0 && response.Code == codes.OK {
		jsonBytes, err := protojson.Marshal(response.Message)
		if err != nil {
			t.Logf("Warning: cannot marshal response for assertions: %v", err)
		} else {
			var responseData map[string]any
			if err := json.Unmarshal(jsonBytes, &responseData); err != nil {
				t.Logf("Warning: cannot parse response for assertions: %v", err)
			} else {
				for _, assertion := range step.GRPCResponse.Assertions {
					renderedAssertion := assertion
					renderedAssertion.Value = RenderAny(assertion.Value, ctxMap)
					if err := AssertResponseV2(renderedAssertion, responseData); err != nil {
						t.Errorf("Assertion failed: %v\n%s", err, formatGRPCDiagnostic(diagnosticStep, response))
					}
				}
			}
		}
	}

	// Execute DB checks
	if len(step.DBChecks) > 0 && db != nil {
		for _, check := range step.DBChecks {
			check = renderDBCheck(check, ctxMap)
			ExecuteStepDBCheck(t, step.Name, db, check)
		}
	}

	return uint64(response.Code)
}

// expandFakerInContext expands faker placeholders in context values
func expandFakerInContext(ctx map[string]any) map[string]any {
	registry := NewFakerRegistry()
	result := make(map[string]any)

	for k, v := range ctx {
		switch val := v.(type) {
		case string:
			result[k] = expandFakerInString(val, registry)
		default:
			result[k] = v
		}
	}

	return result
}

// expandFakerInString replaces {{faker.xxx}} with generated values
func expandFakerInString(input string, registry *FakerRegistry) string {
	// Simple implementation - in production would use regex
	for name := range registry.functions {
		placeholder := fmt.Sprintf("{{faker.%s}}", name)
		if strings.Contains(input, placeholder) {
			generated, _ := registry.Generate(name)
			input = strings.ReplaceAll(input, placeholder, generated)
		}
	}
	return input
}
