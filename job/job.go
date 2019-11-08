package job

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://gcsweb-ci.svc.ci.openshift.org/gcs/origin-ci-test/logs"
)

type Job struct {
	ID         string
	StartedAt  time.Time
	FinishedAt time.Time
	Result     string
}

func Fetch(jobName, jobID string) (Job, error) {
	var client http.Client

	j := Job{ID: jobID}

	// Get start metadata
	{
		req, err := http.NewRequest(http.MethodGet, baseURL+"/"+jobName+"/"+jobID+"/started.json", nil)
		if err != nil {
			return j, err
		}

		res, err := client.Do(req)
		if err != nil {
			return j, err
		}

		var started metadata
		if err := json.NewDecoder(res.Body).Decode(&started); err != nil {
			return j, err
		}

		j.StartedAt = started.time
	}

	// Get finish metadata
	{
		req, err := http.NewRequest(http.MethodGet, baseURL+"/"+jobName+"/"+jobID+"/finished.json", nil)
		if err != nil {
			return j, err
		}

		res, err := client.Do(req)
		if err != nil {
			return j, err
		}

		var finished metadata
		if err := json.NewDecoder(res.Body).Decode(&finished); err != nil {
			return j, err
		}

		j.FinishedAt = finished.time
		j.Result = finished.result
	}

	return j, nil
}

// String implements fmt.Stringer
func (j Job) String() string {
	return strings.Join([]string{
		j.ID,                                   // ID
		j.StartedAt.String(),                   // Started
		j.FinishedAt.Sub(j.StartedAt).String(), // Duration
		j.Result,                               // Result
	}, "\t")

}
