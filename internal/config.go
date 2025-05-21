package internal

type Config struct {
	ConnStr     string
	FixturesDir string

	BeforeReq func() error
	AfterReq  func() error
}
