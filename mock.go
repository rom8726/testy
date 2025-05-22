package testy

import (
	"net/http/httptest"

	"github.com/rom8726/testy/internal"
)

type MockInstance struct {
	Name   string
	Server *httptest.Server
	spy    *internal.SpyStore
}

type MockManager struct {
	Instances []*MockInstance
}

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

func StartMockManager(defs map[string]MockServerDef) (*MockManager, error) {
	var manager MockManager

	for name, def := range defs {
		srv, spy := internal.NewMockServer(name, makeInternalMockServerDef(def))

		inst := &MockInstance{
			Name:   name,
			Server: srv,
			spy:    spy,
		}
		manager.Instances = append(manager.Instances, inst)
	}

	return &manager, nil
}

func (m *MockManager) StopAll() {
	for _, inst := range m.Instances {
		inst.Server.Close()
	}
}

func (m *MockManager) URL(name string) string {
	for _, inst := range m.Instances {
		if inst.Name == name {
			return inst.Server.URL
		}
	}

	return ""
}

func (m *MockManager) InternalInstances() []*internal.MockInstance {
	res := make([]*internal.MockInstance, 0, len(m.Instances))
	for _, inst := range m.Instances {
		res = append(res, internal.NewMockInstance(inst.Name, inst.Server.URL, inst.spy))
	}

	return res
}

func makeInternalMockServerDef(def MockServerDef) internal.MockServerDef {
	routes := make([]internal.MockRoute, 0, len(def.Routes))
	for _, route := range def.Routes {
		routes = append(routes, internal.MockRoute{
			Method: route.Method,
			Path:   route.Path,
			Response: internal.MockResponse{
				Status:  route.Response.Status,
				Headers: route.Response.Headers,
				JSON:    route.Response.JSON,
			},
		})
	}

	return internal.MockServerDef{
		Routes: routes,
	}
}
