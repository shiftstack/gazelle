package job

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/shiftstack/gazelle/pkg/cache"
)

type Job struct {
	Name   string
	Target string
	ID     string

	client http.Client

	cache *cache.Cache
}

func (j Job) baseURL() string {
	return "https://storage.googleapis.com/origin-ci-test/logs/release-openshift-ocp-installer-" + j.Name + "-" + j.Target + "/" + j.ID
}

func (j *Job) fetch(file string) (io.Reader, error) {
	if j.cache == nil {
		j.cache = new(cache.Cache)
	}
	if r, ok := j.cache.Get(file); ok {
		return r, nil
	}

	return j.cache.Fetch(file)
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

func (j Job) BuildLog() (io.Reader, error) {
	return j.fetch(j.baseURL() + "/build-log.txt")
}

func (j Job) Machines() (io.Reader, error) {
	return j.fetch(j.baseURL() + "/artifacts/" + j.Name + "/machines.json")
}

func (j Job) Nodes() (io.Reader, error) {
	return j.fetch(j.baseURL() + "/artifacts/" + j.Name + "/openstack_nodes.log")
}
