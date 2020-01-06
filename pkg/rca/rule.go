package rca

import (
	"bufio"
	"encoding/xml"
	"regexp"

	"github.com/pierreprinetti/go-junit"
)

func infraFailureIfMatchBuildLogs(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
		f, err := j.BuildLog()
		if err != nil {
			return err
		}

		if re.MatchReader(bufio.NewReader(f)) {
			infraFailures <- cause
		}
		return nil
	}
}

func infraFailureIfMatchMachines(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
		f, err := j.Machines()
		if err != nil {
			return err
		}

		if re.MatchReader(bufio.NewReader(f)) {
			infraFailures <- cause
		}
		return nil
	}
}

func infraFailureIfMatchNodes(expr string, cause Cause) Rule {
	re := regexp.MustCompile(expr)
	return func(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
		f, err := j.Nodes()
		if err != nil {
			return err
		}

		if re.MatchReader(bufio.NewReader(f)) {
			infraFailures <- cause
		}
		return nil
	}
}

func failedTests(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	f, err := j.JUnit()
	if err != nil {
		infraFailures <- Cause("Error parsing the JUnit file: " + err.Error())
		return nil
	}

	var testSuite junit.TestSuite
	if err := xml.NewDecoder(f).Decode(&testSuite); err != nil {
		return err
	}

	for _, tc := range testSuite.TestCases {
		if tc.Failure != nil {
			testFailures <- Cause(tc.Name)
		}
	}

	return nil
}
