package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	jobName string
	from    int
	to      int
)

func main() {
	for i := from; i <= to; i++ {
		j := job.Job{
			Name: jobName,
			ID:   strconv.Itoa(i),
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

		rootCause, err := rca.Find(j)
		if err != nil {
			panic(err)
		}

		fmt.Println(strings.Join([]string{
			j.ID,                               // ID
			startedAt.String(),                 // Started
			finishedAt.Sub(startedAt).String(), // Duration
			result,                             // Result
			"",                                 //
			"",                                 //
			"",                                 //
			"",                                 //
			"",                                 //
			"cireport",                         // CI Cop
			rootCause,                          // Root Cause
		}, "\t"))
	}
}

func init() {
	flag.StringVar(&jobName, "job", "", "Name of the test job")
	flag.IntVar(&from, "from", 1, "First job to fetch")
	flag.IntVar(&to, "to", 1, "Last job to fetch")

	flag.Parse()
}
