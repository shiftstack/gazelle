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

func matchBuildLogs(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, failures chan<- Cause) {
		f, err := j.BuildLog()
		if err != nil {
			failures <- CauseGeneric("Failed to get build log: " + err.Error())
			return
		}

		if re.MatchReader(bufio.NewReader(f)) {
			failures <- cause
		}
	}
}

func matchMachines(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, failures chan<- Cause) {
		f, err := j.Machines()
		if err != nil {
			failures <- CauseGeneric("Failed to get Machines information: " + err.Error())
			return
		}

		if re.MatchReader(bufio.NewReader(f)) {
			failures <- cause
		}
	}
}

func matchNodes(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, failures chan<- Cause) {
		f, err := j.Nodes()
		if err != nil {
			failures <- CauseGeneric("Failed to get OpenStack Nodes information: " + err.Error())
			return
		}

		if re.MatchReader(bufio.NewReader(f)) {
			failures <- cause
		}
	}
}

// findBuildLogsInfra creates an infra failure for the first match among the
// given expressions
func findBuildLogsInfra(expressions ...string) Rule {
	re := regexp.MustCompile(strings.Join(expressions, "|"))

	return func(j job, failures chan<- Cause) {
		f, err := j.BuildLog()
		if err != nil {
			failures <- CauseGeneric("Failed to get build log: " + err.Error())
			return
		}

		var buf bytes.Buffer
		r := io.TeeReader(f, &buf)

		if loc := re.FindReaderIndex(bufio.NewReader(r)); loc != nil {
			failures <- CauseInfra(buf.Bytes()[loc[0]:loc[1]])
		}
	}
}

func failedTests(j job, failures chan<- Cause) {
	f, err := j.JUnit()
	if err != nil {
		if err != io.EOF {
			failures <- CauseGeneric("Failed to get the JUnit file: " + err.Error())
		}
		return
	}

	var testSuite junit.TestSuite
	if err := xml.NewDecoder(f).Decode(&testSuite); err != nil {
		failures <- CauseGeneric("Failed to decode the JUnit file: " + err.Error())
		return
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
			return
		}
	}

	for _, tc := range testSuite.TestCases {
		if tc.Failure != nil {
			failures <- CauseGeneric(tc.Name)
		}
	}
}
