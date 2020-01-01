package rca

import (
	"io"
	"io/ioutil"
	"strings"
)

const (
	ErroredVM = "Provisioned VM in ERROR state"
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
