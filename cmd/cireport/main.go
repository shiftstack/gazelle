package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/shiftstack/gazelle/pkg/gsheets"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/prow"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	requestedJobName string
	requestedJobID   int64
	username         string
)

var valid_jobs = []string{
	"release-openshift-ocp-installer-e2e-openstack-4.8",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.8",
	"periodic-ci-openshift-release-master-ci-4.7-e2e-openstack-parallel",
	"periodic-ci-openshift-release-master-ci-4.7-e2e-openstack-serial",
	"release-openshift-ocp-installer-e2e-openstack-4.6",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.6",
	"release-openshift-ocp-installer-e2e-openstack-4.5",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.5",
	"release-openshift-ocp-installer-e2e-openstack-4.4",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.4",
	"release-openshift-ocp-installer-e2e-openstack-4.3",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.3",
}

func toBufferedChannel(numbers ...int64) <-chan int64 {
	ch := make(chan int64, len(numbers))
	for _, n := range numbers {
		ch <- n
	}
	close(ch)
	return ch
}

func main() {
	ctx := context.TODO()
	client := gsheets.NewClient()

	var jobs []string
	if requestedJobName == "" {
		jobs = valid_jobs
	} else {
		jobs = []string{requestedJobName}
	}

	for _, jobName := range jobs {
		fmt.Printf("Updating %s\n", jobName)

		sheet := gsheets.Sheet{
			JobName: jobName,
			Client:  &client,
		}

		var jobIDs <-chan int64
		if requestedJobID != 0 {
			jobIDs = toBufferedChannel(requestedJobID)
		} else {
			lowerBound := sheet.GetLatestId() + 1
			jobIDs = prow.Sorted(prow.JobIDs(ctx, jobName, lowerBound))
		}

		for jobID := range jobIDs {
			fmt.Printf("%s %v\n", jobName, jobID)
			j := job.Job{
				FullName: jobName,
				ID:       strconv.FormatInt(jobID, 10),
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
	flag.StringVar(&requestedJobName, "job", "", "Full name of the test job (e.g. release-openshift-ocp-installer-e2e-openstack-serial-4.4). All known jobs if unset.")
	flag.Int64Var(&requestedJobID, "id", 0, "Job ID. If unset, it consists of all new runs since last time the spreadsheet was updated.")

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
