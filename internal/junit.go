package internal

import (
	"encoding/xml"
	"os"
	"strconv"
	"time"
)

type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Time      string          `xml:"time,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name    string        `xml:"name,attr"`
	Time    string        `xml:"time,attr"`
	Failure *junitFailure `xml:"failure,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
}

func WriteJUnitReport(path, suiteName string, results []TestCaseResult) error {
	suite := junitTestSuite{
		Name:  suiteName,
		Tests: len(results),
	}

	var total time.Duration
	for _, r := range results {
		total += r.Duration

		tc := junitTestCase{
			Name: r.Name,
			Time: formatDuration(r.Duration),
		}
		if r.ErrMsg != "" {
			suite.Failures++
			tc.Failure = &junitFailure{Message: r.ErrMsg}
		}
		suite.TestCases = append(suite.TestCases, tc)
	}
	suite.Time = formatDuration(total)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, _ = f.Write([]byte(xml.Header))
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")

	return enc.Encode(suite)
}

func formatDuration(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 3, 64)
}
