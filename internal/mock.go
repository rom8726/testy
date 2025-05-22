package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/julienschmidt/httprouter"
)

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

func NewMockServer(name string, def MockServerDef) (*httptest.Server, *SpyStore) {
	spy := &SpyStore{
		Calls: new([]MockCall),
	}
	router := httprouter.New()

	for _, route := range def.Routes {
		handler := buildMockHandler(name, route, spy)
		router.Handle(route.Method, route.Path, handler)
	}

	return httptest.NewServer(router), spy
}

func buildMockHandler(name string, route MockRoute, spy *SpyStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Printf(">> mock %q called\n", name)

		body, _ := io.ReadAll(r.Body)

		headers := map[string]string{}
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ", ")
		}

		spy.Add(MockCall{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: headers,
			Body:    string(body),
		})

		for k, v := range route.Response.Headers {
			w.Header().Set(k, v)
		}

		if route.Response.JSON != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(route.Response.Status)
			_, _ = fmt.Fprint(w, route.Response.JSON)

			return
		}

		if route.Response.Body != "" {
			w.WriteHeader(route.Response.Status)
			_, _ = fmt.Fprint(w, route.Response.Body)

			return
		}

		w.WriteHeader(route.Response.Status)
	}
}
