package internal

type Config struct {
	ConnStr     string
	FixturesDir string
	Mocks       []*MockInstance

	BeforeReq func() error
	AfterReq  func() error
}
