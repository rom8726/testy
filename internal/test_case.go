package internal

type TestCase struct {
	Name     string       `yaml:"name"`
	Fixtures []string     `yaml:"fixtures,omitempty"`
	Request  RequestSpec  `yaml:"request"`
	Response ResponseSpec `yaml:"response"`
	DBChecks []DBCheck    `yaml:"dbChecks"`
}

type RequestSpec struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers"`
	Body    any               `yaml:"body"`
}

type ResponseSpec struct {
	Status int    `yaml:"status"`
	JSON   string `yaml:"json"`
}

type DBCheck struct {
	Query  string `yaml:"query"`
	Result any    `yaml:"result"`
}
