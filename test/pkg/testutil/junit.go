package testutil

import (
	"encoding/xml"
	"fmt"
	"os"
	"sync"
	"time"
)

type JUnit struct {
	name      string
	suites    map[string]JUnitTestSuite
	timestamp time.Time
	mutex     sync.Mutex
}

func NewJUnit(name string) *JUnit {
	return &JUnit{
		name:      name,
		suites:    make(map[string]JUnitTestSuite),
		timestamp: time.Now(),
	}
}

func (j *JUnit) AddTest(suite string, t JUnitTestCase) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if s, ok := j.suites[suite]; ok {
		s.TestCases = append(s.TestCases, t)
		s.Time = int(time.Since(s.Timestamp).Seconds())
		j.suites[suite] = s
	} else {
		j.suites[suite] = JUnitTestSuite{
			Name:      suite,
			Timestamp: time.Now(),
			TestCases: []JUnitTestCase{t},
		}
	}
}

func (j *JUnit) WriteToFile(path string) error {
	contents, err := j.ToJUnitTestSuites().Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal junit: %w", err)
	}
	err = os.WriteFile(path, contents, 0644)
	if err != nil {
		return fmt.Errorf("failed to write junit to file %q: %w", path, err)
	}
	return nil
}

func (j *JUnit) ToJUnitTestSuites() *JUnitTestSuites {
	suites := &JUnitTestSuites{
		Name:      j.name,
		Timestamp: j.timestamp,
		Time:      int(time.Since(j.timestamp).Seconds()),
	}

	for _, suite := range j.suites {
		suites.TestSuites = append(suites.TestSuites, suite)
	}

	return suites
}

type JUnitTestSuites struct {
	XMLName    xml.Name  `xml:"testsuites"`
	Name       string    `xml:"name,attr,omitempty"`
	Time       int       `xml:"time,attr"`
	Timestamp  time.Time `xml:"timestamp,attr,omitempty"`
	TestSuites []JUnitTestSuite
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
