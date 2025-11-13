package internal

import (
	"database/sql"
	"fmt"
	"net/http"
)

// HookType defines the type of hook
type HookType string

const (
	HookTypeSetup    HookType = "setup"
	HookTypeTeardown HookType = "teardown"
)

// Hook represents a setup or teardown action
type Hook struct {
	SQL  string        `yaml:"sql,omitempty"`  // SQL query to execute
	HTTP *HTTPHookSpec `yaml:"http,omitempty"` // HTTP request to make
	Name string        `yaml:"name,omitempty"` // Optional name for logging
}

// HTTPHookSpec defines an HTTP request for hooks
type HTTPHookSpec struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    any               `yaml:"body,omitempty"`
}

// HookExecutor executes setup and teardown hooks
type HookExecutor struct {
	db      *sql.DB
	baseURL string
}

// NewHookExecutor creates a new hook executor
func NewHookExecutor(db *sql.DB, baseURL string) *HookExecutor {
	return &HookExecutor{
		db:      db,
		baseURL: baseURL,
	}
}

// ExecuteHooks executes a list of hooks
func (e *HookExecutor) ExecuteHooks(hooks []Hook, hookType HookType, ctx map[string]any) error {
	for i, hook := range hooks {
		hookName := hook.Name
		if hookName == "" {
			hookName = fmt.Sprintf("%s hook #%d", hookType, i+1)
		}

		if err := e.executeHook(hook, hookName, ctx); err != nil {
			return fmt.Errorf("%s failed: %w", hookName, err)
		}
	}

	return nil
}

// executeHook executes a single hook
func (e *HookExecutor) executeHook(hook Hook, name string, ctx map[string]any) error {
	// SQL hook
	if hook.SQL != "" {
		return e.executeSQLHook(hook.SQL, ctx)
	}

	// HTTP hook
	if hook.HTTP != nil {
		return e.executeHTTPHook(hook.HTTP, ctx)
	}

	return fmt.Errorf("hook has no action defined")
}

// executeSQLHook executes a SQL hook
func (e *HookExecutor) executeSQLHook(query string, ctx map[string]any) error {
	if e.db == nil {
		return fmt.Errorf("database connection not available for SQL hook")
	}

	// Render the query with context
	renderedQuery := RenderTemplate(query, ctx)

	_, err := e.db.Exec(renderedQuery)
	if err != nil {
		return fmt.Errorf("SQL execution failed: %w", err)
	}

	return nil
}

// executeHTTPHook executes an HTTP hook
func (e *HookExecutor) executeHTTPHook(spec *HTTPHookSpec, ctx map[string]any) error {
	// Render the request
	renderedPath := RenderTemplate(spec.Path, ctx)

	// Create HTTP request
	req, err := http.NewRequest(spec.Method, e.baseURL+renderedPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	for k, v := range spec.Headers {
		renderedValue := RenderTemplate(v, ctx)
		req.Header.Set(k, renderedValue)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP request returned error status: %d", resp.StatusCode)
	}

	return nil
}

// TestCaseHooks contains setup and teardown hooks for a test case
type TestCaseHooks struct {
	Setup    []Hook `yaml:"setup,omitempty"`
	Teardown []Hook `yaml:"teardown,omitempty"`
}

// StepHooks contains setup and teardown hooks for a single step
type StepHooks struct {
	Before []Hook `yaml:"before,omitempty"`
	After  []Hook `yaml:"after,omitempty"`
}

// ExecuteSetup executes all setup hooks for a test case
func ExecuteSetup(executor *HookExecutor, hooks []Hook, ctx map[string]any) error {
	return executor.ExecuteHooks(hooks, HookTypeSetup, ctx)
}

// ExecuteTeardown executes all teardown hooks for a test case
func ExecuteTeardown(executor *HookExecutor, hooks []Hook, ctx map[string]any) error {
	return executor.ExecuteHooks(hooks, HookTypeTeardown, ctx)
}

// DeferredTeardown wraps teardown execution to ensure it runs even on panic
type DeferredTeardown struct {
	executor *HookExecutor
	hooks    []Hook
	ctx      map[string]any
	executed bool
}

// NewDeferredTeardown creates a new deferred teardown
func NewDeferredTeardown(executor *HookExecutor, hooks []Hook, ctx map[string]any) *DeferredTeardown {
	return &DeferredTeardown{
		executor: executor,
		hooks:    hooks,
		ctx:      ctx,
		executed: false,
	}
}

// Execute runs the teardown hooks
func (d *DeferredTeardown) Execute() error {
	if d.executed {
		return nil
	}

	d.executed = true
	return ExecuteTeardown(d.executor, d.hooks, d.ctx)
}

// Defer should be called with defer to ensure teardown runs
func (d *DeferredTeardown) Defer() {
	if err := d.Execute(); err != nil {
		// Log error but don't fail the test
		fmt.Printf("Warning: teardown failed: %v\n", err)
	}
}
