package testy

import (
	"net/http"
	"testing"

	"github.com/rom8726/testy/internal"
)

type Config struct {
	Handler     http.Handler
	CasesDir    string
	FixturesDir string
	ConnStr     string
}

func Run(t *testing.T, cfg *Config) {
	t.Helper()

	cases, err := internal.LoadTestCases(cfg.CasesDir)
	if err != nil {
		t.Fatalf("cannot load test cases: %v", err)
	}

	for _, tc := range cases {
		cfgInternal := internal.Config{
			ConnStr:     cfg.ConnStr,
			FixturesDir: cfg.FixturesDir,
		}

		internal.RunSingle(t, cfg.Handler, tc, &cfgInternal)
	}
}
