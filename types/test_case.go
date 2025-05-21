package types

type TestCase struct {
	Name     string       `yaml:"name"`
	Request  RequestSpec  `yaml:"request"`
	Response ResponseSpec `yaml:"response"`
}

type RequestSpec struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers"`
	Body    any               `yaml:"body"`
}

type ResponseSpec struct {
	Status int            `yaml:"status"`
	JSON   map[string]any `yaml:"json"`
}
