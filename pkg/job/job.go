package job

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/shiftstack/gazelle/pkg/cache"
)

const (
	baseURL = "https://gcsweb-ci.svc.ci.openshift.org/gcs/origin-ci-test/logs"
)

type Job struct {
	Name string
	ID   string

	client http.Client

	cache *cache.Cache
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
	f, err := j.fetch(baseURL + "/" + j.Name + "/" + j.ID + "/started.json")
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
	f, err := j.fetch(baseURL + "/" + j.Name + "/" + j.ID + "/finished.json")
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
	f, err := j.fetch(baseURL + "/" + j.Name + "/" + j.ID + "/finished.json")
	if err != nil {
		return "", err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return "", err
	}

	return finished.result, nil
}
