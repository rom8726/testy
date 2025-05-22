package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/rom8726/testy"

	"testyexample"
)

func TestServer(t *testing.T) {
	connStr := "postgresql://user:password@localhost:5432/db?sslmode=disable"

	mocks, err := testy.StartMockManager(map[string]testy.MockServerDef{
		"notification": {
			Routes: []testy.MockRoute{
				{
					Method: "POST",
					Path:   "/send",
					Response: testy.MockResponse{
						Status: http.StatusAccepted,
						JSON:   `{"status":"queued"}`,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("mock start: %v", err)
	}
	defer mocks.StopAll()

	err = os.Setenv("NOTIFICATION_BASE_URL", mocks.URL("notification"))
	if err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("NOTIFICATION_BASE_URL")

	srv := testyexample.NewServer(connStr)

	cfg := testy.Config{
		Handler:     srv.Router,
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
	}

	testy.Run(t, &cfg)
}
