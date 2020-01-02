package rca

import (
	"io"
)

var (
	rules = []func(job) (string, bool, error){
		erroredMachine,
		erroredNode,
		erroredVolume,
	}
)

type job interface {
	Result() (string, error)
	BuildLog() (io.Reader, error)
	Machines() (io.Reader, error)
	Nodes() (io.Reader, error)
}

func Find(j job) (string, error) {
	res, err := j.Result()
	if err != nil {
		return "", err
	}
	if res == "SUCCESS" {
		return "", nil
	}

	for _, apply := range rules {
		rc, ok, err := apply(j)
		if err != nil {
			return "", err
		}
		if ok {
			return rc, nil
		}
	}

	return "", nil
}
