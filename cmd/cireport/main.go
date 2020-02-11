package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/pierreprinetti/go-sequence"
	"github.com/shiftstack/gazelle/pkg/gsheets"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/prow"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	fullJobName string
	jobIDs      string
	username    string
)

func main() {
	client := gsheets.NewClient()
	sheet  := gsheets.Sheet {
		JobName: fullJobName,
		Client:  &client,
	}

	if jobIDs == "" {
		lowerBound := sheet.GetLatestId() + 1
		upperBound := prow.GetLatestId(fullJobName)
		jobIDs = fmt.Sprintf("%d-%d", lowerBound, upperBound)
	}

	ids, err := sequence.Int(jobIDs)
	if err != nil {
		panic(err)
	}

	for _, i := range ids {
		j := job.Job{
			FullName: fullJobName,
			UserName: username,
			ID:       strconv.Itoa(i),
		}

		result, err := j.Result()
		if err == nil {
			j.ComputedResult = result
		} else {
			j.ComputedResult = "Pending"
		}

		var (
			testFailures  []string
			infraFailures []string
		)
		for failure := range rca.Find(j) {
			if failure.IsInfra() {
				infraFailures = append(infraFailures, failure.String())
			}
			testFailures = append(testFailures, failure.String())
		}

		j.RootCause = testFailures
		if len(infraFailures) > 0 {
			j.RootCause = infraFailures
			j.ComputedResult = "INFRA FAILURE"
		}

		sheet.AddRow(j)
	}
}

func init() {
	flag.StringVar(&fullJobName, "job", "", "Full name of the test job (e.g. release-openshift-ocp-installer-e2e-openstack-serial-4.4)")
	flag.StringVar(&jobIDs, "id", "", "Job IDs")

	flag.StringVar(&username, "user", os.Getenv("CIREPORT_USER"), "Username to use for CI Cop")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		} else {
			username = "cireport"
		}
	}

	flag.Parse()
}
