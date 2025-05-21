package testy

import (
	"net/http"
	"testing"

	"github.com/rom8726/testy/loader"
	"github.com/rom8726/testy/runner"
)

func Run(t *testing.T, handler http.Handler, testsDir string) {
	cases, err := loader.LoadTestCases(testsDir)
	if err != nil {
		t.Fatalf("cannot load test cases: %v", err)
	}

	for _, tc := range cases {
		runner.RunSingle(t, handler, tc)
	}
}
