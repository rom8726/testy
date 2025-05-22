package internal

type MockRoute struct {
	Method   string       `yaml:"method"`
	Path     string       `yaml:"path"`
	Response MockResponse `yaml:"response"`
}

type MockResponse struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers"`
	JSON    string            `yaml:"json"`
	Body    string            `yaml:"body"`
}

type MockServerDef struct {
	Routes []MockRoute `yaml:"routes"`
}

type MockCall struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type SpyStore struct {
	Calls *[]MockCall
}

func (s *SpyStore) Add(call MockCall) {
	*s.Calls = append(*s.Calls, call)
}
