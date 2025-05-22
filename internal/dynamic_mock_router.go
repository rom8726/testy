package internal

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type DynamicMockRouter struct {
	name   string
	router *httprouter.Router
	spy    *SpyStore
}

func NewDynamicMockRouter(name string) *DynamicMockRouter {
	return &DynamicMockRouter{
		name:   name,
		router: httprouter.New(),
		spy:    &SpyStore{Calls: new([]MockCall)},
	}
}

func (d *DynamicMockRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.router.ServeHTTP(w, r)
}

func (d *DynamicMockRouter) AddRoute(route MockRoute) {
	d.router.Handle(route.Method, route.Path, buildMockHandler(d.name, route, d.spy))
}

func (d *DynamicMockRouter) Spy() *SpyStore {
	return d.spy
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
