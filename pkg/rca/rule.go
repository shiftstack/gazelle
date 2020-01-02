package rca

import (
	"encoding/xml"
	"io/ioutil"
	"strings"

	"github.com/pierreprinetti/go-junit"
)

type RuleFunc func(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error

func (f RuleFunc) Apply(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	return f(j, testFailures, infraFailures)
}

func erroredMachine(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	logs, err := j.Machines()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(logs)
	if err != nil {
		return err
	}

	if strings.Contains(string(b), `"machine.openshift.io/instance-state": "ERROR"`) {
		infraFailures <- CauseErroredVM
	}

	return nil
}

func erroredNode(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	logs, err := j.Nodes()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(logs)
	if err != nil {
		return err
	}

	if strings.Contains(string(b), "ERROR") {
		infraFailures <- CauseErroredVM
	}

	return nil
}

func erroredVolume(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	logs, err := j.BuildLog()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(logs)
	if err != nil {
		return err
	}

	if strings.Contains(string(b), "The volume is in error status. Please check with your cloud admin") {
		infraFailures <- CauseErroredVolume
	}

	return nil
}

func failedTests(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	f, err := j.JUnit()
	if err != nil {
		testFailures <- Cause("Error parsing the JUnit file: " + err.Error())
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

func machineTimeout(j job, testFailures chan<- Cause, infraFailures chan<- Cause) error {
	logs, err := j.BuildLog()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(logs)
	if err != nil {
		return err
	}

	if strings.Contains(string(b), "timeout while waiting for state to become 'ACTIVE' (last state: 'BUILD', timeout: 30m0s)") {
		infraFailures <- CauseMachineTimeout
	}

	return nil
}
