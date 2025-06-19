package internal

type TestCase struct {
	Name      string                   `yaml:"name"`
	Fixtures  []string                 `yaml:"fixtures,omitempty"`
	Mocks     map[string]MockServerDef `yaml:"mockServers,omitempty"`
	MockCalls []MockCallCheck          `yaml:"mockCalls,omitempty"`
	Steps     []Step                   `yaml:"steps"`
}

type Step struct {
	Name     string       `yaml:"name"`
	Request  RequestSpec  `yaml:"request"`
	Response ResponseSpec `yaml:"response"`
	DBChecks []DBCheck    `yaml:"dbChecks"`
}

type RequestSpec struct {
	Method   string            `yaml:"method"`
	Path     string            `yaml:"path"`
	Headers  map[string]string `yaml:"headers"`
	Body     any               `yaml:"body"`
	BodyFile string            `yaml:"bodyFile"`
	BodyRaw  string            `yaml:"bodyRaw"`
}

type ResponseSpec struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers"`
	JSON    string            `yaml:"json"`
	Text    string            `yaml:"text"`
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
