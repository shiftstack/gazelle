package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pierreprinetti/go-sequence"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	fullJobName string
	jobIDs  string
)

func main() {
	ids, err := sequence.Int(jobIDs)
	if err != nil {
		panic(err)
	}

	for _, i := range ids {
		j := job.Job{
			FullName: fullJobName,
			ID:     strconv.Itoa(i),
		}

		startedAt, err := j.StartTime()
		if err != nil {
			panic(err)
		}

		finishedAt, err := j.FinishTime()
		if err != nil {
			panic(err)
		}

		result, err := j.Result()
		if err != nil {
			panic(err)
		}

		failures, errs := rca.Find(j)

		var wg sync.WaitGroup
		wg.Add(1)
		// panic at the first error
		go func() {
			for err := range errs {
				panic(err)
			}
			wg.Done()
		}()

		var (
			testFailures  []string
			infraFailures []string
		)
		for failure := range failures {
			if failure.IsInfra() {
				infraFailures = append(infraFailures, failure.String())
			}
			testFailures = append(testFailures, failure.String())
		}

		// Wait for the error handling to occur
		wg.Wait()

		rootCause := testFailures
		if len(infraFailures) > 0 {
			rootCause = infraFailures
			result = "INFRA FAILURE"
		}

		var machinesURL string
		machinesURL, err = j.MachinesURL()
		if err != nil {
			panic(err)
		}
		var nodesURL string
		nodesURL, err = j.NodesURL()
		if err != nil {
			panic(err)
		}

		var s strings.Builder
		{
			s.WriteString(`<meta http-equiv="content-type" content="text/html; charset=utf-8"><meta name="generator" content="cireport"/><table xmlns="http://www.w3.org/1999/xhtml"><tbody><tr><td>`)
			s.WriteString(strings.Join([]string{
				`<a href="` + j.JobURL() + `">` + j.ID + `</a>`,
				startedAt.String(),
				finishedAt.Sub(startedAt).String(),
				result,
				"",
				`<a href="` + j.BuildLogURL() + `">` + j.BuildLogURL() + `</a>`,
				`<a href="` + machinesURL + `">` + machinesURL + `</a>`,
				`<a href="` + nodesURL + `">` + nodesURL + `</a>`,
				"cireport",
				strings.Join(rootCause, "<br />"),
			}, "</td><td>"))
			s.WriteString(`</td></tr></tbody></table>`)
		}
		fmt.Println(s.String())
	}
}

func init() {
	flag.StringVar(&fullJobName, "job", "", "Full name of the test job (e.g. release-openshift-ocp-installer-e2e-openstack-serial-4.4)")
	flag.StringVar(&jobIDs, "id", "", "Job IDs")

	flag.Parse()
}
