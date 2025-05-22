package internal

type MockInstance struct {
	name string
	url  string
	spy  *SpyStore
}

func NewMockInstance(name, url string, spy *SpyStore) *MockInstance {
	return &MockInstance{
		name: name,
		url:  url,
		spy:  spy,
	}
}
