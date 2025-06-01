package testy

import (
	"net/http"
	"testing"

	"github.com/rom8726/pgfixtures"

	"github.com/rom8726/testy/internal"
)

type Config struct {
	Handler     http.Handler
	DBType      pgfixtures.DatabaseType
	CasesDir    string
	FixturesDir string
	ConnStr     string
	MockManager *MockManager

	BeforeReq func() error
	AfterReq  func() error

	JUnitReport string
}

func Run(t *testing.T, cfg *Config) {
	t.Helper()

	cases, err := internal.LoadTestCases(cfg.CasesDir)
	if err != nil {
		// Use the error directly since it's already wrapped by LoadTestCases
		t.Fatalf("%v", err)
	}

	var mocks []*internal.MockInstance
	if cfg.MockManager != nil {
		mocks = cfg.MockManager.internalInstances()
	}

	results := make([]internal.TestCaseResult, 0, len(cases))

	for _, tc := range cases {
		cfgInternal := internal.Config{
			DBType:      cfg.DBType,
			ConnStr:     cfg.ConnStr,
			FixturesDir: cfg.FixturesDir,
			Mocks:       mocks,
			BeforeReq:   cfg.BeforeReq,
			AfterReq:    cfg.AfterReq,
		}

		res := internal.RunSingle(t, cfg.Handler, tc, &cfgInternal)
		results = append(results, res)
	}

	if cfg.JUnitReport != "" {
		if err := internal.WriteJUnitReport(cfg.JUnitReport, "testy", results); err != nil {
			// Use the error directly since it's already wrapped by WriteJUnitReport
			// Using Logf instead of Fatalf to allow tests to continue
			t.Logf("%v", err)
		}
	}
}
