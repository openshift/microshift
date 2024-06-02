package util

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"
)

type JUnitTestSuites struct {
	XMLName    xml.Name  `xml:"testsuites"`
	Name       string    `xml:"name,attr,omitempty"`
	Time       int       `xml:"time,attr"`
	Timestamp  time.Time `xml:"timestamp,attr,omitempty"`
	TestSuites []JUnitTestSuite
}

func (j *JUnitTestSuites) WriteToFile(path string) error {
	contents, err := j.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal junit: %w", err)
	}
	err = os.WriteFile(path, contents, 0644)
	if err != nil {
		return fmt.Errorf("failed to write junit to file %q: %w", path, err)
	}
	return nil
}

// Prow's junit lens does not support nested TestSuites
// https://github.com/k8s-ci-robot/test-infra/blob/d7867ec05b41/prow/spyglass/lenses/junit/lens.go#L107

type JUnitTestSuite struct {
	XMLName   xml.Name  `xml:"testsuite"`
	Name      string    `xml:"name,attr,omitempty"`
	Time      int       `xml:"time,attr"`
	Timestamp time.Time `xml:"timestamp,attr,omitempty"`
	Tests     int       `xml:"tests,attr,omitempty"`
	Failures  int       `xml:"failures,attr,omitempty"`
	Skipped   int       `xml:"skipped,attr,omitempty"`
	TestCases []JUnitTestCase
}

type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr,omitempty"`
	ClassName string        `xml:"classname,attr,omitempty"`
	Time      int           `xml:"time,attr,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
	Skipped   *JUnitSkipped `xml:"skipped,omitempty"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitSkipped struct {
	Message string `xml:"message,attr"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Content string `xml:",chardata"`
}

func (suites *JUnitTestSuites) Marshal() ([]byte, error) {
	suites.Time = int(time.Since(suites.Timestamp).Seconds())
	bs, err := xml.MarshalIndent(suites, "", "    ")
	if err != nil {
		return nil, err
	}
	return bs, nil
}
