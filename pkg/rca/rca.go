package rca

import (
	"io"
	"sync"
)

var (
	rules = []Rule{
		matchBuildLogs(
			"to become ready: unexpected state 'ERROR', wanted target 'ACTIVE'. last error",
			CauseErroredVM,
		),

		matchBuildLogs(
			"to become ready: timeout while waiting for state to become 'ACTIVE'",
			CauseErroredVM,
		),

		matchBuildLogs(
			"The volume is in error status. Please check with your cloud admin",
			CauseErroredVolume,
		),

		matchBuildLogs(
			"Cluster operator authentication Progressing is True with ProgressingWellKnownNotReady: Progressing: got '404 Not Found' status while trying to GET the OAuth well-known",
			CauseClusterTimeout,
		),

		matchBuildLogs(
			"failed to initialize the cluster: Cluster operator [\\w-]+ is still updating",
			CauseClusterTimeout,
		),

		matchBuildLogs(
			"failed to initialize the cluster: Working towards",
			CauseClusterTimeout,
		),

		matchBuildLogs(
			"failed to initialize the cluster: Multiple errors are preventing progress",
			CauseClusterTimeout,
		),

		matchBuildLogs(
			"failed to wait for bootstrapping to complete",
			CauseBootstrapTimeout,
		),

		matchBuildLogs(
			"failed: unable to import latest release image",
			CauseReleaseImage,
		),

		matchBuildLogs(
			"failed: unable to find the '[\\w-]+' image in the provided release image",
			CauseReleaseImage,
		),

		matchMachines(
			`"machine.openshift.io/instance-state": "ERROR"`,
			CauseErroredVM,
		),

		matchNodes(
			"ERROR",
			CauseErroredVM,
		),

		matchBuildLogs(
			`Quota exceeded for resources: \['router'\]`,
			CauseQuota("router"),
		),

		matchBuildLogs(
			"VolumeSizeExceedsAvailableQuota: Requested volume or snapshot exceeds allowed gigabytes quota",
			CauseQuota("volume size"),
		),

		matchBuildLogs(
			`when calling the ChangeResourceRecordSets operation`,
			CauseRoute53,
		),

		matchBuildLogs(
			`failed to acquire lease`,
			CauseLeaseFailure,
		),

		findBuildLogsInfra(
			`error: could not run steps: step \[release-inputs\] failed.*`,
			`An unexpected error prevented the server from fulfilling your request. \(HTTP \d{3}\)`,
			`error: could not run steps: step \[release:latest\] failed: the following tags from the release could not be imported to stable after five minutes.*`,
		),

		failedTests,
	}
)

type Rule func(j job, failures chan<- Cause)

type job interface {
	Result() (string, error)
	BuildLog() (io.Reader, error)
	Machines() (io.Reader, error)
	Nodes() (io.Reader, error)
	JUnit() (io.Reader, error)
}

func Find(j job) <-chan Cause {
	failures := make(chan Cause)

	res, _ := j.Result()
	if res == "SUCCESS" {
		close(failures)
		return failures
	}

	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(rule Rule) {
			rule(j, failures)
			wg.Done()
		}(rule)
	}

	go func() {
		wg.Wait()
		close(failures)
	}()

	return uniqueFilter(failures)
}

func uniqueFilter(inCh <-chan Cause) <-chan Cause {
	var (
		outCh = make(chan Cause)
		cache = make(map[Cause]struct{})
	)

	go func() {
		for cause := range inCh {
			if _, ok := cache[cause]; ok {
				continue
			}
			outCh <- cause
			cache[cause] = struct{}{}
		}
		close(outCh)
	}()

	return outCh
}
