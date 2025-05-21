package tests

import (
	"fmt"
	"testing"

	"github.com/rom8726/testy"
	"testyexample"
)

func TestServer(t *testing.T) {
	connStr := "postgresql://user:password@localhost:5432/db?sslmode=disable"

	srv := testyexample.NewServer(connStr)

	cfg := testy.Config{
		Handler:     srv.Router,
		CasesDir:    "tests/cases",
		FixturesDir: "tests/fixtures",
		ConnStr:     connStr,
		BeforeReq: func() error {
			fmt.Println("before request")

			return nil
		},
		AfterReq: func() error {
			fmt.Println("after request")

			return nil
		},
	}

	testy.Run(t, &cfg)
}
