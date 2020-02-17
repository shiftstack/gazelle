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

var valid_jobs = []string{
	"release-openshift-ocp-installer-e2e-openstack-4.4",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.4",
	"release-openshift-ocp-installer-e2e-openstack-4.3",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.3",
	"release-openshift-ocp-installer-e2e-openstack-4.2",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.2",
	"release-openshift-origin-installer-e2e-openstack-4.2",
	"release-openshift-origin-installer-e2e-openstack-serial-4.2",
}

func main() {
	client := gsheets.NewClient()

	var jobs []string
	if fullJobName == "" {
		jobs = valid_jobs
	} else {
		jobs = append(jobs, fullJobName)
	}

	for _, jobName := range jobs {
		sheet := gsheets.Sheet{
			JobName: jobName,
			Client:  &client,
		}

		realJobIDs := jobIDs
		if realJobIDs == "" {
			lowerBound := sheet.GetLatestId() + 1
			upperBound := prow.GetLatestId(jobName)
			if lowerBound < upperBound {
				realJobIDs = fmt.Sprintf("%d-%d", lowerBound, upperBound)
			} else if lowerBound == upperBound {
				realJobIDs = fmt.Sprintf("%d", upperBound)
			} else {
				// This sheet is already up-to-date
				continue
			}
		}
		fmt.Printf("Updating %s with results from jobs %s\n", jobName, realJobIDs)

		ids, err := sequence.Int(realJobIDs)
		if err != nil {
			panic(err)
		}

		for _, i := range ids {
			fmt.Printf("%s %v\n", jobName, i)
			j := job.Job{
				FullName: jobName,
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

			sheet.AddRow(j, username)
		}
	}
}

func init() {
	flag.StringVar(&fullJobName, "job", "", "Full name of the test job (e.g. release-openshift-ocp-installer-e2e-openstack-serial-4.4). All known jobs if unset.")
	flag.StringVar(&jobIDs, "id", "", "Job IDs. If unset, it consists of all new runs since last time the spreadsheet was updated.")

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
