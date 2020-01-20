package job

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/shiftstack/gazelle/pkg/cache"
)

var jobTargetRexexp = regexp.MustCompile(`release-openshift-(ocp|origin)-installer-(.*)-\d.\d`)

type Job struct {
	FullName string
	Target   string
	ID       string

	client http.Client

	cache *cache.Cache
}

func (j Job) baseURL() string {
	return "https://storage.googleapis.com/origin-ci-test/logs/" + j.FullName + "/" + j.ID
}

func (j *Job) fetch(file string) (io.Reader, error) {
	if j.cache == nil {
		j.cache = new(cache.Cache)
	}
	return j.cache.Get(file)
}

func (j *Job) Name() (string, error) {
	matches := jobTargetRexexp.FindStringSubmatch(j.FullName)

	if len(matches) >= 3 {
		return matches[2], nil
	} else {
		return "", fmt.Errorf("Could not determine job name from %s", j.FullName)
	}
}

func (j Job) StartTime() (time.Time, error) {
	f, err := j.fetch(j.baseURL() + "/started.json")
	if err != nil {
		return time.Time{}, err
	}

	var started metadata
	if err := json.NewDecoder(f).Decode(&started); err != nil {
		return time.Time{}, err
	}

	return started.time, nil
}

func (j Job) FinishTime() (time.Time, error) {
	f, err := j.fetch(j.baseURL() + "/finished.json")
	if err != nil {
		return time.Time{}, err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return time.Time{}, err
	}

	return finished.time, nil
}

func (j Job) Result() (string, error) {
	f, err := j.fetch(j.baseURL() + "/finished.json")
	if err != nil {
		return "", err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return "", err
	}

	return finished.result, nil
}

func (j Job) BuildLogURL() string {
	return j.baseURL() + "/build-log.txt"
}

func (j Job) BuildLog() (io.Reader, error) {
	return j.fetch(j.BuildLogURL())
}

func (j Job) MachinesURL() (string, error) {
	name, err := j.Name()
	if err != nil {
		return "", err
	}
	return j.baseURL() + "/artifacts/" + name + "/machines.json", nil
}

func (j Job) Machines() (io.Reader, error) {
	url, err := j.MachinesURL()
	if err != nil {
		return nil, err
	}
	return j.fetch(url)
}

func (j Job) NodesURL() (string, error) {
	name, err := j.Name()
	if err != nil {
		return "", err
	}
	return j.baseURL() + "/artifacts/" + name + "/openstack_nodes.log", nil
}

func (j Job) Nodes() (io.Reader, error) {
	url, err := j.NodesURL()
	if err != nil {
		return nil, err
	}
	return j.fetch(url)
}

func (j Job) JobURL() string {
	return "https://prow.svc.ci.openshift.org/view/gcs/origin-ci-test/logs/" + j.FullName + "/" + j.ID
}

func (j Job) JUnitURL() (string, error) {
	const (
		target       = "Writing JUnit report to /tmp/artifacts/junit/"
		targetLength = len(target)
	)

	buildLog, err := j.BuildLog()
	if err != nil {
		return "", err
	}

	// The default scanner buffer (64*1024 bytes) is too short for some
	// build logs. This sets the initial buffer capacity to the package
	// default, but a higher maximum value.
	scanner := bufio.NewScanner(buildLog)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var jobName string
	jobName, err = j.Name()
	if err != nil {
		return "", err
	}
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) >= targetLength && line[:len(target)] == target {
			filename := line[len(target):]
			return j.baseURL() + "/artifacts/" + jobName + "/junit/" + filename, scanner.Err()
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", io.EOF
}

func (j Job) JUnit() (io.Reader, error) {
	u, err := j.JUnitURL()
	if err != nil {
		return nil, err
	}
	return j.fetch(u)
}
