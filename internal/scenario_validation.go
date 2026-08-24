package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ValidationConfig contains runtime capabilities needed for static scenario validation.
type ValidationConfig struct {
	CasesDir    string
	FixturesDir string
	HasHandler  bool
	GRPCAddr    string
	Mocks       []*MockInstance
}

// ScenarioValidationError aggregates all scenario authoring problems found before execution.
type ScenarioValidationError struct {
	Problems []string
}

func (e *ScenarioValidationError) Error() string {
	if len(e.Problems) == 0 {
		return "scenario validation failed"
	}

	var b strings.Builder
	b.WriteString("scenario validation failed")
	for _, problem := range e.Problems {
		b.WriteString("; ")
		b.WriteString(problem)
	}

	return b.String()
}

// ValidateTestCases validates Testy YAML cases before any scenario is executed.
func ValidateTestCases(cases []TestCase, cfg ValidationConfig) error {
	v := scenarioValidator{
		cfg:       cfg,
		problems:  make([]string, 0),
		operators: knownAssertionOperators(),
		grpcCodes: knownGRPCCodes(),
		fakers:    NewFakerRegistry(),
	}

	for i, tc := range cases {
		v.validateCase(i, tc)
	}

	if len(v.problems) > 0 {
		return &ScenarioValidationError{Problems: v.problems}
	}

	return nil
}

type scenarioValidator struct {
	cfg       ValidationConfig
	problems  []string
	operators map[string]struct{}
	grpcCodes map[string]struct{}
	fakers    *FakerRegistry
}

func (v *scenarioValidator) validateCase(index int, tc TestCase) {
	caseRef := fmt.Sprintf("case[%d]", index)
	if strings.TrimSpace(tc.Name) != "" {
		caseRef = fmt.Sprintf("%s (%q)", caseRef, tc.Name)
	}

	if strings.TrimSpace(tc.Name) == "" {
		v.add("%s: name is required", caseRef)
	}
	if len(tc.Steps) == 0 {
		v.add("%s: steps must contain at least one step", caseRef)
	}

	v.validateFixtures(caseRef, tc.Fixtures)
	v.validateMockServers(caseRef, tc.Mocks)
	v.validateMockCalls(caseRef, tc.MockCalls, tc.Mocks)

	variableNames := make(map[string]struct{}, len(tc.Variables))
	for name := range tc.Variables {
		variableNames[name] = struct{}{}
	}

	previousSteps := make(map[string]struct{}, len(tc.Steps))
	allSteps := make(map[string]int, len(tc.Steps))
	for i, step := range tc.Steps {
		normalized := normalizeStepName(step.Name)
		if normalized == "" {
			continue
		}
		if firstIndex, exists := allSteps[normalized]; exists {
			v.add("%s step[%d] (%q): duplicate step name after normalization; first seen at step[%d]", caseRef, i, step.Name, firstIndex)
			continue
		}
		allSteps[normalized] = i
	}

	for i, step := range tc.Steps {
		stepRef := fmt.Sprintf("%s step[%d]", caseRef, i)
		if strings.TrimSpace(step.Name) != "" {
			stepRef = fmt.Sprintf("%s (%q)", stepRef, step.Name)
		}

		loopVars := v.validateStep(stepRef, step)
		allowedNames := copyStringSet(variableNames)
		for name := range previousSteps {
			allowedNames[name] = struct{}{}
		}
		if normalized := normalizeStepName(step.Name); normalized != "" {
			allowedNames[normalized] = struct{}{}
		}
		for _, name := range loopVars {
			allowedNames[name] = struct{}{}
		}

		v.validatePlaceholders(stepRef, step, allowedNames, allSteps, previousSteps)

		normalized := normalizeStepName(step.Name)
		if normalized != "" {
			previousSteps[normalized] = struct{}{}
		}
	}
}

func (v *scenarioValidator) validateStep(stepRef string, step Step) []string {
	var loopVars []string

	if strings.TrimSpace(step.Name) == "" {
		v.add("%s: name is required", stepRef)
	}

	isHTTP := step.Request != nil
	isGRPC := step.GRPCRequest != nil
	if !isHTTP && !isGRPC {
		v.add("%s: either request or grpcRequest is required", stepRef)
	}
	if isHTTP && isGRPC {
		v.add("%s: request and grpcRequest are mutually exclusive", stepRef)
	}
	if step.Response != nil && !isHTTP {
		v.add("%s: response requires request", stepRef)
	}
	if step.GRPCResponse != nil && !isGRPC {
		v.add("%s: grpcResponse requires grpcRequest", stepRef)
	}

	if isHTTP {
		if !v.cfg.HasHandler {
			v.add("%s: HTTP request requires Config.Handler", stepRef)
		}
		if step.Response == nil {
			v.add("%s: HTTP request requires response", stepRef)
		}
		v.validateHTTPRequest(stepRef, step.Request)
		v.validateHTTPResponse(stepRef, step.Response)
	}

	if isGRPC {
		if v.cfg.GRPCAddr == "" {
			v.add("%s: gRPC request requires Config.GRPCAddr", stepRef)
		}
		if step.GRPCResponse == nil {
			v.add("%s: gRPC request requires grpcResponse", stepRef)
		}
		v.validateGRPCRequest(stepRef, step.GRPCRequest)
		v.validateGRPCResponse(stepRef, step.GRPCResponse)
	}

	if step.Loop != nil {
		loopVars = v.validateLoop(stepRef, step.Loop)
	}
	if step.Retry != nil {
		if _, err := ParseRetryConfig(*step.Retry); err != nil {
			v.add("%s retry: %v", stepRef, err)
		}
		for _, code := range step.Retry.RetryOn {
			if !validHTTPStatus(code) {
				v.add("%s retry.retryOn: status code %d is outside 100..599", stepRef, code)
			}
		}
	}
	if step.Performance != nil {
		if _, err := ParsePerformanceSpec(*step.Performance); err != nil {
			v.add("%s performance: %v", stepRef, err)
		}
	}

	for j, check := range step.DBChecks {
		checkRef := fmt.Sprintf("%s dbChecks[%d]", stepRef, j)
		if strings.TrimSpace(check.Query) == "" {
			v.add("%s: query is required", checkRef)
		}
		if check.Result == nil {
			v.add("%s: result is required", checkRef)
		}
	}

	return loopVars
}

func (v *scenarioValidator) validateHTTPRequest(stepRef string, req *RequestSpec) {
	if req == nil {
		return
	}
	if strings.TrimSpace(req.Method) == "" {
		v.add("%s request: method is required", stepRef)
	}
	if strings.TrimSpace(req.Path) == "" {
		v.add("%s request: path is required", stepRef)
	}
	bodyFields := 0
	if req.Body != nil {
		bodyFields++
	}
	if req.BodyFile != "" {
		bodyFields++
		if !pathExists(req.BodyFile) {
			v.add("%s request.bodyFile: file does not exist: %s", stepRef, req.BodyFile)
		}
	}
	if req.BodyRaw != "" {
		bodyFields++
	}
	if bodyFields > 1 {
		v.add("%s request: use only one of body, bodyFile, or bodyRaw", stepRef)
	}
}

func (v *scenarioValidator) validateHTTPResponse(stepRef string, resp *ResponseSpec) {
	if resp == nil {
		return
	}
	if !validHTTPStatus(resp.Status) {
		v.add("%s response.status: status code %d is outside 100..599", stepRef, resp.Status)
	}
	if resp.Schema != "" {
		if !pathExists(resp.Schema) {
			v.add("%s response.schema: file does not exist: %s", stepRef, resp.Schema)
		} else if _, err := LoadJSONSchemaFromFile(resp.Schema); err != nil {
			v.add("%s response.schema: %v", stepRef, err)
		}
	}
	v.validateAssertions(stepRef+" response", resp.Assertions)
}

func (v *scenarioValidator) validateGRPCRequest(stepRef string, req *GRPCRequestSpec) {
	if req == nil {
		return
	}
	if strings.TrimSpace(req.Service) == "" {
		v.add("%s grpcRequest: service is required", stepRef)
	}
	if strings.TrimSpace(req.Method) == "" {
		v.add("%s grpcRequest: method is required", stepRef)
	}
	if req.Message == nil {
		v.add("%s grpcRequest: message is required", stepRef)
	}
}

func (v *scenarioValidator) validateGRPCResponse(stepRef string, resp *GRPCResponseSpec) {
	if resp == nil {
		return
	}
	if _, ok := v.grpcCodes[resp.Code]; !ok {
		v.add("%s grpcResponse.code: unknown gRPC status code %q", stepRef, resp.Code)
	}
	v.validateAssertions(stepRef+" grpcResponse", resp.Assertions)
}

func (v *scenarioValidator) validateAssertions(ref string, assertions []ResponseAssertion) {
	for i, assertion := range assertions {
		assertRef := fmt.Sprintf("%s assertions[%d]", ref, i)
		if strings.TrimSpace(assertion.Path) == "" {
			v.add("%s: path is required", assertRef)
		}
		if _, ok := v.operators[assertion.Operator]; !ok {
			v.add("%s: unknown operator %q", assertRef, assertion.Operator)
		}
	}
}

func (v *scenarioValidator) validateLoop(stepRef string, loop *LoopConfig) []string {
	var vars []string
	if strings.TrimSpace(loop.Var) == "" {
		v.add("%s loop: var is required", stepRef)
	} else {
		vars = append(vars, loop.Var)
	}

	hasItems := len(loop.Items) > 0
	hasRange := loop.Range != nil
	if hasItems == hasRange {
		v.add("%s loop: exactly one of items or range is required", stepRef)
	}
	if loop.Range != nil {
		if loop.Range.Step == 0 {
			v.add("%s loop.range: step cannot be 0", stepRef)
		}
		if loop.Range.Step > 0 && loop.Range.From > loop.Range.To {
			v.add("%s loop.range: from must be <= to for a positive step", stepRef)
		}
		if loop.Range.Step < 0 && loop.Range.From < loop.Range.To {
			v.add("%s loop.range: from must be >= to for a negative step", stepRef)
		}
	}

	return vars
}

func (v *scenarioValidator) validateFixtures(caseRef string, fixtures []string) {
	if len(fixtures) == 0 || v.cfg.FixturesDir == "" {
		return
	}
	for _, fixture := range fixtures {
		if strings.TrimSpace(fixture) == "" {
			v.add("%s fixtures: fixture name cannot be empty", caseRef)
			continue
		}
		if !fixtureExists(v.cfg.FixturesDir, fixture) {
			v.add("%s fixtures: fixture file does not exist for %q in %s", caseRef, fixture, v.cfg.FixturesDir)
		}
	}
}

func (v *scenarioValidator) validateMockServers(caseRef string, mocks map[string]MockServerDef) {
	for name, def := range mocks {
		mockRef := fmt.Sprintf("%s mockServers[%q]", caseRef, name)
		if strings.TrimSpace(name) == "" {
			v.add("%s: mock name cannot be empty", mockRef)
		}
		if FindMockInstance(v.cfg.Mocks, name) == nil {
			v.add("%s: mock is not present in MockManager", mockRef)
		}
		if len(def.Routes) == 0 {
			v.add("%s: routes must contain at least one route", mockRef)
		}
		for i, route := range def.Routes {
			routeRef := fmt.Sprintf("%s routes[%d]", mockRef, i)
			if strings.TrimSpace(route.Method) == "" {
				v.add("%s: method is required", routeRef)
			}
			if strings.TrimSpace(route.Path) == "" {
				v.add("%s: path is required", routeRef)
			}
			if !validHTTPStatus(route.Response.Status) {
				v.add("%s response.status: status code %d is outside 100..599", routeRef, route.Response.Status)
			}
		}
	}
}

func (v *scenarioValidator) validateMockCalls(caseRef string, checks []MockCallCheck, mockDefs map[string]MockServerDef) {
	for i, check := range checks {
		checkRef := fmt.Sprintf("%s mockCalls[%d]", caseRef, i)
		if strings.TrimSpace(check.Mock) == "" {
			v.add("%s: mock is required", checkRef)
			continue
		}
		_, definedInCase := mockDefs[check.Mock]
		if !definedInCase && FindMockInstance(v.cfg.Mocks, check.Mock) == nil {
			v.add("%s: mock %q is neither declared in mockServers nor present in MockManager", checkRef, check.Mock)
		}
		if check.Count < 0 {
			v.add("%s: count cannot be negative", checkRef)
		}
	}
}

func (v *scenarioValidator) validatePlaceholders(
	ref string,
	value any,
	allowedNames map[string]struct{},
	allSteps map[string]int,
	previousSteps map[string]struct{},
) {
	for _, placeholder := range collectPlaceholders(value) {
		if strings.HasPrefix(placeholder, "faker.") {
			name := strings.TrimPrefix(placeholder, "faker.")
			if _, err := v.fakers.Generate(name); err != nil {
				v.add("%s: unknown faker placeholder {{%s}}", ref, placeholder)
			}
			continue
		}

		root := placeholderRoot(placeholder)
		if _, ok := allowedNames[root]; ok {
			continue
		}
		if _, ok := allSteps[root]; ok {
			if _, wasExecutedBefore := previousSteps[root]; !wasExecutedBefore {
				v.add("%s: placeholder {{%s}} references step %q before it is available", ref, placeholder, root)
			}
			continue
		}

		// Environment variables are dynamic by design, so only validate known DSL references.
		if strings.Contains(placeholder, ".response") {
			v.add("%s: placeholder {{%s}} references unknown step or variable %q", ref, placeholder, root)
		}
	}
}

func (v *scenarioValidator) add(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

func collectPlaceholders(value any) []string {
	var placeholders []string
	var visit func(any)

	visit = func(v any) {
		switch val := v.(type) {
		case string:
			matches := placeholderRe.FindAllStringSubmatch(val, -1)
			for _, match := range matches {
				if len(match) > 1 {
					placeholders = append(placeholders, match[1])
				}
			}
		case map[string]any:
			keys := make([]string, 0, len(val))
			for key := range val {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				visit(val[key])
			}
		case []any:
			for _, item := range val {
				visit(item)
			}
		case TestCase:
			visit(val.Variables)
			visit(val.Mocks)
			visit(val.MockCalls)
			visit(val.Setup)
			visit(val.Teardown)
			visit(val.Steps)
		case Step:
			visit(val.When)
			visit(val.Request)
			visit(val.Response)
			visit(val.GRPCRequest)
			visit(val.GRPCResponse)
			visit(val.Performance)
			visit(val.DBChecks)
		case *RequestSpec:
			if val != nil {
				visit(*val)
			}
		case RequestSpec:
			visit(val.Path)
			visit(val.Headers)
			visit(val.Body)
			visit(val.BodyRaw)
			visit(val.BodyFile)
		case *ResponseSpec:
			if val != nil {
				visit(*val)
			}
		case ResponseSpec:
			visit(val.Headers)
			visit(val.JSON)
			visit(val.Text)
			visit(val.Schema)
			visit(val.Assertions)
		case ResponseAssertion:
			visit(val.Path)
			visit(val.Value)
			visit(val.Message)
		case *GRPCRequestSpec:
			if val != nil {
				visit(*val)
			}
		case GRPCRequestSpec:
			visit(val.Service)
			visit(val.Method)
			visit(val.Message)
			visit(val.Metadata)
		case *GRPCResponseSpec:
			if val != nil {
				visit(*val)
			}
		case GRPCResponseSpec:
			visit(val.Message)
			visit(val.Metadata)
			visit(val.Assertions)
		case DBCheck:
			visit(val.Query)
			visit(val.Result)
		case []DBCheck:
			for _, item := range val {
				visit(item)
			}
		case []ResponseAssertion:
			for _, item := range val {
				visit(item)
			}
		case map[string]string:
			keys := make([]string, 0, len(val))
			for key := range val {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				visit(val[key])
			}
		}
	}

	visit(value)
	return placeholders
}

func copyStringSet(in map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeStepName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}

func placeholderRoot(placeholder string) string {
	if idx := strings.IndexAny(placeholder, ".["); idx >= 0 {
		return placeholder[:idx]
	}
	return placeholder
}

func validHTTPStatus(status int) bool {
	return status >= 100 && status <= 599
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fixtureExists(fixturesDir, fixture string) bool {
	candidates := []string{filepath.Join(fixturesDir, fixture)}
	if filepath.Ext(fixture) == "" {
		candidates = append(candidates,
			filepath.Join(fixturesDir, fixture+".yml"),
			filepath.Join(fixturesDir, fixture+".yaml"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return true
		} else if !errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return false
}

func knownAssertionOperators() map[string]struct{} {
	operators := []string{
		"equals", "eq", "==",
		"notEquals", "ne", "!=",
		"greaterThan", "gt", ">",
		"lessThan", "lt", "<",
		"greaterOrEqual", "gte", ">=",
		"lessOrEqual", "lte", "<=",
		"contains", "notContains", "matches",
		"startsWith", "endsWith", "between",
		"in", "notIn", "isEmpty", "isNotEmpty",
		"hasLength", "hasMinLength", "hasMaxLength",
	}
	out := make(map[string]struct{}, len(operators))
	for _, op := range operators {
		out[op] = struct{}{}
	}
	return out
}

func knownGRPCCodes() map[string]struct{} {
	codes := []string{
		"OK",
		"CANCELLED", "Cancelled",
		"UNKNOWN", "Unknown",
		"INVALID_ARGUMENT", "InvalidArgument",
		"DEADLINE_EXCEEDED", "DeadlineExceeded",
		"NOT_FOUND", "NotFound",
		"ALREADY_EXISTS", "AlreadyExists",
		"PERMISSION_DENIED", "PermissionDenied",
		"RESOURCE_EXHAUSTED", "ResourceExhausted",
		"FAILED_PRECONDITION", "FailedPrecondition",
		"ABORTED", "Aborted",
		"OUT_OF_RANGE", "OutOfRange",
		"UNIMPLEMENTED", "Unimplemented",
		"INTERNAL", "Internal",
		"UNAVAILABLE", "Unavailable",
		"DATA_LOSS", "DataLoss",
		"UNAUTHENTICATED", "Unauthenticated",
	}
	out := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		out[code] = struct{}{}
	}
	return out
}
