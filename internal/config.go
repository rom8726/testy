package internal

import (
	"github.com/rom8726/pgfixtures"
)

// Config holds the configuration for running tests
type Config struct {
	DBType      pgfixtures.DatabaseType
	ConnStr     string
	FixturesDir string
	Mocks       []*MockInstance

	BeforeReq func() error
	AfterReq  func() error
}
