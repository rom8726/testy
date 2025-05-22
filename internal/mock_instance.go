package internal

type MockInstance struct {
	name   string
	url    string
	router *DynamicMockRouter
}

func NewMockInstance(name, url string, router *DynamicMockRouter) *MockInstance {
	return &MockInstance{
		name:   name,
		url:    url,
		router: router,
	}
}
