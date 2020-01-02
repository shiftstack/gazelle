package rca

import (
	"io/ioutil"
	"strings"
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
