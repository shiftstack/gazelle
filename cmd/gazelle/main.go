package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/shiftstack/gazelle/pkg/job"
)

func main() {
	const (
		jobName   = "release-openshift-ocp-installer-e2e-openstack-serial-4.5"
		runNumber = "173"
	)

	j := job.Job{
		FullName: jobName,
		ID:       runNumber,
	}

	started, err := j.StartTime()
	if err != nil {
		panic(err)
	}

	result, err := j.Result()
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(Run{
		JobName:       j.FullName,
		RunNumber:     j.ID,
		Started:       started,
		Duration:      j.Duration().String(),
		Result:        result,
		BuildLogURL:   j.BuildLogURL(),
		MatchingRules: []string{},
	}); err != nil {
		log.Fatal(err)
	}
}
