package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/rom8726/pgfixtures"

	"github.com/rom8726/testy/v2"

	"testyexample"
)

func TestServer(t *testing.T) {
	mocks, err := testy.StartMockManager("notification")
	if err != nil {
		t.Fatalf("mock start: %v", err)
	}
	defer mocks.StopAll()

	err = os.Setenv("NOTIFICATION_BASE_URL", mocks.URL("notification"))
	if err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("NOTIFICATION_BASE_URL")

	connStr := "postgresql://user:password@localhost:5432/db?sslmode=disable"
	srv := testyexample.NewServer(connStr)

	cfg := testy.Config{
		Handler:     srv.Router,
		DBType:      pgfixtures.PostgreSQL,
		CasesDir:    "./cases",
		FixturesDir: "./fixtures",
		ConnStr:     connStr,
		MockManager: mocks,
		BeforeReq: func() error {
			fmt.Println("before request")

			return nil
		},
		AfterReq: func() error {
			fmt.Println("after request")

			return nil
		},
		JUnitReport: "./junit.xml",
	}

	testy.Run(t, &cfg)
}
