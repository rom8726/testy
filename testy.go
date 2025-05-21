package testy

import (
	"net/http"
	"testing"

	"github.com/rom8726/testy/internal"
)

func Run(t *testing.T, handler http.Handler, testsDir string) {
	cases, err := internal.LoadTestCases(testsDir)
	if err != nil {
		t.Fatalf("cannot load test cases: %v", err)
	}

	for _, tc := range cases {
		internal.RunSingle(t, handler, tc)
	}
}
