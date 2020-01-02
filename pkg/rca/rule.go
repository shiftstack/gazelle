package rca

import (
	"io"
	"io/ioutil"
	"strings"
)

const (
	ErroredVM     = "Provisioned VM in ERROR state"
	ErroredVolume = "Provisioned Volume in ERROR state"
)

func erroredMachine(j job) (rc string, ok bool, err error) {
	var (
		logs io.Reader
		b    []byte
	)

	logs, err = j.Machines()
	if err != nil {
		return
	}

	b, err = ioutil.ReadAll(logs)
	if err != nil {
		return
	}

	if strings.Contains(string(b), `"machine.openshift.io/instance-state": "ERROR"`) {
		rc = ErroredVM
		ok = true
	}

	return
}

func erroredNode(j job) (rc string, ok bool, err error) {
	var (
		logs io.Reader
		b    []byte
	)

	logs, err = j.Nodes()
	if err != nil {
		return
	}

	b, err = ioutil.ReadAll(logs)
	if err != nil {
		return
	}

	if strings.Contains(string(b), "ERROR") {
		rc = ErroredVM
		ok = true
	}

	return
}

func erroredVolume(j job) (rc string, ok bool, err error) {
	var (
		logs io.Reader
		b    []byte
	)

	logs, err = j.BuildLog()
	if err != nil {
		return
	}

	b, err = ioutil.ReadAll(logs)
	if err != nil {
		return
	}

	if strings.Contains(string(b), "The volume is in error status. Please check with your cloud admin") {
		rc = ErroredVolume
		ok = true
	}

	return
}
