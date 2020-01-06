package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	jobName string
	target  string
	from    int
	to      int
)

func main() {
	for i := from; i <= to; i++ {
		j := job.Job{
			Name:   jobName,
			Target: target,
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

		testFailuresCh, infraFailuresCh, errs := rca.Find(j)

		// panic at the first error
		go func() {
			for err := range errs {
				panic(err)
			}
		}()

		var wg sync.WaitGroup
		wg.Add(2)
		var (
			testFailures  []string
			infraFailures []string
		)
		go func() {
			testFailures = reduceFailures(testFailuresCh)
			wg.Done()
		}()
		go func() {
			infraFailures = reduceFailures(infraFailuresCh)
			wg.Done()
		}()
		wg.Wait()

		rootCause := testFailures
		if len(infraFailures) > 0 {
			rootCause = infraFailures
			result = "INFRA FAILURE"
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
				`<a href="` + j.MachinesURL() + `">` + j.MachinesURL() + `</a>`,
				`<a href="` + j.NodesURL() + `">` + j.NodesURL() + `</a>`,
				"cireport",
				strings.Join(rootCause, "<br />"),
			}, "</td><td>"))
			s.WriteString(`</td></tr></tbody></table>`)
		}
		fmt.Println(s.String())
	}
}

func reduceFailures(failures <-chan rca.Cause) []string {
	var out []string
	for failure := range failures {
		out = append(out, string(failure))
	}
	return out
}

func init() {
	flag.StringVar(&jobName, "job", "", "Name of the test job")
	flag.StringVar(&target, "target", "", "Target OpenShift version")
	flag.IntVar(&from, "from", 1, "First job to fetch")
	flag.IntVar(&to, "to", 1, "Last job to fetch")

	flag.Parse()
}
