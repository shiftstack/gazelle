package rca

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"strings"

	"github.com/pierreprinetti/go-junit"
)

func ifMatchBuildLogs(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, failures chan<- Cause) error {
		f, err := j.BuildLog()
		if err != nil {
			failures <- CauseGeneric("Failed to get build log: " + err.Error())
			return nil
		}

		if re.MatchReader(bufio.NewReader(f)) {
			failures <- cause
		}
		return nil
	}
}

func ifMatchMachines(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, failures chan<- Cause) error {
		f, err := j.Machines()
		if err != nil {
			failures <- CauseGeneric("Failed to get Machines information: " + err.Error())
			return nil
		}

		if re.MatchReader(bufio.NewReader(f)) {
			failures <- cause
		}
		return nil
	}
}

func ifMatchNodes(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, failures chan<- Cause) error {
		f, err := j.Nodes()
		if err != nil {
			failures <- CauseGeneric("Failed to get OpenStack Nodes information: " + err.Error())
			return nil
		}

		if re.MatchReader(bufio.NewReader(f)) {
			failures <- cause
		}
		return nil
	}
}

// findBuildLogsInfra creates an infra failure for the first match among the
// given expressions
func findBuildLogsInfra(expressions ...string) Rule {
	re := regexp.MustCompile(strings.Join(expressions, "|"))

	return func(j job, failures chan<- Cause) error {
		f, err := j.BuildLog()
		if err != nil {
			failures <- CauseGeneric("Failed to get build log: " + err.Error())
			return nil
		}

		var buf bytes.Buffer
		r := io.TeeReader(f, &buf)

		if loc := re.FindReaderIndex(bufio.NewReader(r)); loc != nil {
			failures <- CauseInfra(buf.Bytes()[loc[0]:loc[1]])
		}

		return nil
	}
}

func failedTests(j job, failures chan<- Cause) error {
	f, err := j.JUnit()
	if err != nil {
		if err != io.EOF {
			failures <- CauseGeneric("Error parsing the JUnit file: " + err.Error())
		}
		return nil
	}

	var testSuite junit.TestSuite
	if err := xml.NewDecoder(f).Decode(&testSuite); err != nil {
		return err
	}

	// FIXME(mandre) ideally, we should read this value from the `failures`
	// field of the junit file, however it currently contains bogus values.
	total_failures := 0
	for _, tc := range testSuite.TestCases {
		if tc.Failure != nil {
			total_failures++
		}
		if total_failures > 10 {
			failures <- CauseGeneric("More than 10 failed tests")
			return nil
		}
	}

	for _, tc := range testSuite.TestCases {
		if tc.Failure != nil {
			failures <- CauseGeneric(tc.Name)
		}
	}

	return nil
}
