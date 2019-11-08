package main

import (
	"fmt"
	"strconv"

	"github.com/shiftstack-dev-tools/gazelle/job"
)

const (
	jobName = "release-openshift-ocp-installer-e2e-openstack-serial-4.3"
	from    = 338
	to      = 348
)

func fromTo(from, to int) []string {
	jobIDs := make([]string, to-from+1)
	for i := range jobIDs {
		jobIDs[to-from-i] = strconv.Itoa(from + i)
	}

	return jobIDs
}

func main() {
	for _, jobID := range fromTo(from, to) {

		j, err := job.Fetch(jobName, jobID)
		if err != nil {
			panic(err)
		}

		fmt.Println(j)
	}
}
