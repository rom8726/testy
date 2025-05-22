package testy

import (
	"net/http/httptest"

	"github.com/rom8726/testy/internal"
)

type MockInstance struct {
	name   string
	server *httptest.Server
	router *internal.DynamicMockRouter
}

type MockManager struct {
	instances []*MockInstance
}

func StartMockManager(names ...string) (*MockManager, error) {
	var manager MockManager

	for _, name := range names {
		router := internal.NewDynamicMockRouter(name)
		srv := httptest.NewServer(router)

		inst := &MockInstance{
			name:   name,
			server: srv,
			router: router,
		}
		manager.instances = append(manager.instances, inst)
	}

	return &manager, nil
}

func (m *MockManager) StopAll() {
	for _, inst := range m.instances {
		inst.server.Close()
	}
}

func (m *MockManager) URL(name string) string {
	for _, inst := range m.instances {
		if inst.name == name {
			return inst.server.URL
		}
	}

	return ""
}

func (m *MockManager) internalInstances() []*internal.MockInstance {
	res := make([]*internal.MockInstance, 0, len(m.instances))
	for _, inst := range m.instances {
		res = append(res, internal.NewMockInstance(inst.name, inst.server.URL, inst.router))
	}

	return res
}
