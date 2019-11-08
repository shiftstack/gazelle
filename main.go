package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/shiftstack-dev-tools/gazelle/job"
)

var (
	jobName string
	from    int
	to      int
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

func init() {
	flag.StringVar(&jobName, "job", "", "Name of the test job")
	flag.IntVar(&from, "from", 1, "First job to fetch")
	flag.IntVar(&to, "to", 1, "Last job to fetch")

	flag.Parse()
}
