package rca

import (
	"io"
	"sync"
)

var (
	rules = []Rule{
		ifMatchBuildLogs(
			"to become ready: unexpected state 'ERROR', wanted target 'ACTIVE'. last error",
			CauseErroredVM,
		),

		ifMatchBuildLogs(
			"to become ready: timeout while waiting for state to become 'ACTIVE'",
			CauseErroredVM,
		),

		ifMatchBuildLogs(
			"The volume is in error status. Please check with your cloud admin",
			CauseErroredVolume,
		),

		ifMatchBuildLogs(
			"Cluster operator authentication Progressing is True with ProgressingWellKnownNotReady: Progressing: got '404 Not Found' status while trying to GET the OAuth well-known",
			CauseClusterTimeout,
		),

		ifMatchBuildLogs(
			"failed to initialize the cluster: Cluster operator [\\w-]+ is still updating",
			CauseClusterTimeout,
		),

		ifMatchBuildLogs(
			"failed to initialize the cluster: Working towards",
			CauseClusterTimeout,
		),

		ifMatchBuildLogs(
			"failed to initialize the cluster: Multiple errors are preventing progress",
			CauseClusterTimeout,
		),

		ifMatchBuildLogs(
			"failed: unable to import latest release image",
			CauseReleaseImage,
		),

		ifMatchBuildLogs(
			"failed: unable to find the '[\\w-]+' image in the provided release image",
			CauseReleaseImage,
		),

		ifMatchMachines(
			`"machine.openshift.io/instance-state": "ERROR"`,
			CauseErroredVM,
		),

		ifMatchNodes(
			"ERROR",
			CauseErroredVM,
		),

		ifMatchBuildLogs(
			`Quota exceeded for resources: \['router'\]`,
			CauseQuota("router"),
		),

		ifMatchBuildLogs(
			"VolumeSizeExceedsAvailableQuota: Requested volume or snapshot exceeds allowed gigabytes quota",
			CauseQuota("volume size"),
		),

		ifMatchBuildLogs(
			`when calling the ChangeResourceRecordSets operation`,
			CauseRoute53,
		),

		ifMatchBuildLogs(
			`failed to acquire lease`,
			CauseLeaseFailure,
		),

		failedTests,
	}
)

type Rule func(j job, failures chan<- Cause) error

type job interface {
	Result() (string, error)
	BuildLog() (io.Reader, error)
	Machines() (io.Reader, error)
	Nodes() (io.Reader, error)
	JUnit() (io.Reader, error)
}

func Find(j job) (<-chan Cause, <-chan error) {

	failures := make(chan Cause)
	errs := make(chan error, len(rules))

	res, _ := j.Result()
	if res == "SUCCESS" {
		close(failures)
		close(errs)
		return failures, errs
	}

	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(rule Rule) {
			if err := rule(j, failures); err != nil {
				errs <- err
			}
			wg.Done()
		}(rule)
	}

	go func() {
		wg.Wait()
		close(failures)
		close(errs)
	}()

	return uniqueFilter(failures), errs
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
