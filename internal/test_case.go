package internal

type TestCase struct {
	Name      string                   `yaml:"name"`
	Variables map[string]any           `yaml:"variables,omitempty"`
	Fixtures  []string                 `yaml:"fixtures,omitempty"`
	Mocks     map[string]MockServerDef `yaml:"mockServers,omitempty"`
	MockCalls []MockCallCheck          `yaml:"mockCalls,omitempty"`
	Setup     []Hook                   `yaml:"setup,omitempty"`
	Teardown  []Hook                   `yaml:"teardown,omitempty"`
	Steps     []Step                   `yaml:"steps"`
}

type Step struct {
	Name        string           `yaml:"name"`
	When        string           `yaml:"when,omitempty"`
	Loop        *LoopConfig      `yaml:"loop,omitempty"`
	Retry       *RetryConfig     `yaml:"retry,omitempty"`
	Request     RequestSpec      `yaml:"request"`
	Response    ResponseSpec     `yaml:"response"`
	Performance *PerformanceSpec `yaml:"performance,omitempty"`
	DBChecks    []DBCheck        `yaml:"dbChecks,omitempty"`
}

type LoopConfig struct {
	Items []any  `yaml:"items"`
	Var   string `yaml:"var"`
	Range *Range `yaml:"range,omitempty"`
}

type Range struct {
	From int `yaml:"from"`
	To   int `yaml:"to"`
	Step int `yaml:"step,omitempty"`
}

type RequestSpec struct {
	Method   string            `yaml:"method"`
	Path     string            `yaml:"path"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Body     any               `yaml:"body,omitempty"`
	BodyFile string            `yaml:"bodyFile,omitempty"`
	BodyRaw  string            `yaml:"bodyRaw,omitempty"`
}

type ResponseSpec struct {
	Status     int                 `yaml:"status"`
	Headers    map[string]string   `yaml:"headers,omitempty"`
	JSON       string              `yaml:"json,omitempty"`
	Text       string              `yaml:"text,omitempty"`
	Schema     string              `yaml:"schema,omitempty"`
	JSONSchema *JSONSchema         `yaml:"jsonSchema,omitempty"`
	Assertions []ResponseAssertion `yaml:"assertions,omitempty"`
}

type ResponseAssertion struct {
	Path     string `yaml:"path"`
	Operator string `yaml:"operator"`
	Value    any    `yaml:"value,omitempty"`
	Message  string `yaml:"message,omitempty"`
}

type DBCheck struct {
	Query  string `yaml:"query"`
	Result any    `yaml:"result"`
}

type MockCallExpect struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Body   struct {
		Contains string `yaml:"contains"`
	} `yaml:"body"`
}

type MockCallCheck struct {
	Mock   string         `yaml:"mock"`
	Count  int            `yaml:"count"`
	Expect MockCallExpect `yaml:"expect"`
}
