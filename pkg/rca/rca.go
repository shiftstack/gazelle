package rca

import (
	"io"
	"sync"
)

var (
	rules = []Rule{
		RuleFunc(erroredMachine),
		RuleFunc(erroredNode),
		RuleFunc(erroredVolume),
		RuleFunc(failedTests),
	}
)

type Rule interface {
	Apply(j job, testFailure chan<- Cause, infraFailure chan<- Cause) error
}

type job interface {
	Result() (string, error)
	BuildLog() (io.Reader, error)
	Machines() (io.Reader, error)
	Nodes() (io.Reader, error)
	JUnit() (io.Reader, error)
}

func Find(j job) (<-chan Cause, <-chan Cause, <-chan error) {

	testFailures := make(chan Cause)
	infraFailures := make(chan Cause)
	errs := make(chan error, len(rules))

	res, _ := j.Result()
	if res == "SUCCESS" {
		close(testFailures)
		close(infraFailures)
		close(errs)
		return testFailures, infraFailures, errs
	}

	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(rule Rule) {
			if err := rule.Apply(j, testFailures, infraFailures); err != nil {
				errs <- err
			}
			wg.Done()
		}(rule)
	}

	go func() {
		wg.Wait()
		close(testFailures)
		close(infraFailures)
		close(errs)
	}()

	return testFailures, uniqueFilter(infraFailures), errs
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
