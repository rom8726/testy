package internal

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"google.golang.org/protobuf/encoding/protojson"
)

const diagnosticValueLimit = 4096

func formatHTTPDiagnostic(step Step, rec *httptest.ResponseRecorder) string {
	var b strings.Builder

	if step.Name != "" {
		fmt.Fprintf(&b, "step: %s\n", step.Name)
	}
	if step.Request != nil {
		fmt.Fprintf(&b, "request: %s %s\n", step.Request.Method, step.Request.Path)
		if len(step.Request.Headers) > 0 {
			fmt.Fprintf(&b, "request headers: %s\n", formatStringMap(step.Request.Headers))
		}
		if body := formatRequestBody(*step.Request); body != "" {
			fmt.Fprintf(&b, "request body: %s\n", body)
		}
	}
	if rec != nil {
		fmt.Fprintf(&b, "response status: %d\n", rec.Result().StatusCode)
		if len(rec.Header()) > 0 {
			fmt.Fprintf(&b, "response headers: %s\n", formatHeaderMap(rec.Header()))
		}
		fmt.Fprintf(&b, "response body: %s", truncateDiagnosticValue(strings.TrimSpace(rec.Body.String())))
	}

	return strings.TrimSpace(b.String())
}

func formatDBDiagnostic(stepName string, check DBCheck, actual string) string {
	var b strings.Builder

	if stepName != "" {
		fmt.Fprintf(&b, "step: %s\n", stepName)
	}
	fmt.Fprintf(&b, "query: %s\n", strings.TrimSpace(check.Query))
	fmt.Fprintf(&b, "expected rows: %s\n", truncateDiagnosticValue(formatAny(check.Result)))
	fmt.Fprintf(&b, "actual rows: %s", truncateDiagnosticValue(actual))

	return strings.TrimSpace(b.String())
}

func formatGRPCDiagnostic(step Step, response *GRPCResponse) string {
	var b strings.Builder

	if step.Name != "" {
		fmt.Fprintf(&b, "step: %s\n", step.Name)
	}
	if step.GRPCRequest != nil {
		fmt.Fprintf(&b, "grpc request: %s/%s\n", step.GRPCRequest.Service, step.GRPCRequest.Method)
		if len(step.GRPCRequest.Metadata) > 0 {
			fmt.Fprintf(&b, "request metadata: %s\n", formatStringMap(step.GRPCRequest.Metadata))
		}
		if step.GRPCRequest.Message != nil {
			fmt.Fprintf(&b, "request message: %s\n", truncateDiagnosticValue(formatAny(step.GRPCRequest.Message)))
		}
	}
	if response != nil {
		fmt.Fprintf(&b, "response code: %s\n", response.Code.String())
		if len(response.Metadata) > 0 {
			fmt.Fprintf(&b, "response metadata: %s\n", formatHeaderMap(response.Metadata))
		}
		if response.Error != nil {
			fmt.Fprintf(&b, "response error: %s\n", response.Error.Error())
		}
		if response.Message != nil {
			data, err := protojson.Marshal(response.Message)
			if err != nil {
				fmt.Fprintf(&b, "response message: <marshal failed: %v>", err)
			} else {
				fmt.Fprintf(&b, "response message: %s", truncateDiagnosticValue(string(data)))
			}
		}
	}

	return strings.TrimSpace(b.String())
}

func formatRequestBody(req RequestSpec) string {
	switch {
	case req.BodyFile != "":
		return "@" + req.BodyFile
	case req.BodyRaw != "":
		return truncateDiagnosticValue(req.BodyRaw)
	case req.Body != nil:
		return truncateDiagnosticValue(formatAny(req.Body))
	default:
		return ""
	}
}

func renderStepForDiagnostics(step Step, ctx map[string]any) Step {
	if step.Request != nil {
		rendered := renderRequest(*step.Request, ctx)
		step.Request = &rendered
	}
	if step.GRPCRequest != nil {
		rendered := *step.GRPCRequest
		rendered.Service = RenderTemplate(rendered.Service, ctx)
		rendered.Method = RenderTemplate(rendered.Method, ctx)
		if rendered.Message != nil {
			if msg, ok := RenderAny(rendered.Message, ctx).(map[string]any); ok {
				rendered.Message = msg
			}
		}
		for k, v := range rendered.Metadata {
			rendered.Metadata[k] = RenderTemplate(v, ctx)
		}
		step.GRPCRequest = &rendered
	}

	return step
}

func formatAny(v any) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%+v", val)
		}
		return string(data)
	}
}

func formatStringMap(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", key, values[key]))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func formatHeaderMap(values map[string][]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", key, strings.Join(values[key], ",")))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func truncateDiagnosticValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= diagnosticValueLimit {
		return value
	}

	return value[:diagnosticValueLimit] + fmt.Sprintf("... [truncated, %d bytes total]", len(value))
}

func assertJSONWithDiagnostic(t *testing.T, actual, expected, diagnosticTitle, diagnostic string) {
	t.Helper()

	printer := &diagnosticJSONPrinter{t: t}
	ja := jsonassert.New(printer)
	ja.Assert(actual, expected)

	if printer.failed {
		t.Errorf("%s:\n%s", diagnosticTitle, diagnostic)
	}
}

type diagnosticJSONPrinter struct {
	t      *testing.T
	failed bool
}

func (p *diagnosticJSONPrinter) Errorf(format string, args ...interface{}) {
	p.t.Helper()
	p.failed = true
	p.t.Errorf(format, args...)
}

func (p *diagnosticJSONPrinter) Helper() {
	p.t.Helper()
}
