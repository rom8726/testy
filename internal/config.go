package internal

// Config holds the configuration for running tests
type Config struct {
	ConnStr     string
	FixturesDir string
	Mocks       []*MockInstance

	BeforeReq func() error
	AfterReq  func() error
}
