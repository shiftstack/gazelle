package prow

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/shiftstack/gazelle/pkg/job"
)

type watcher struct {
	sync.RWMutex
	// maps the job names with their most recent known JobID
	jobs map[string]int64

	Jobs <-chan job.Job
}

func (w *watcher) setLastKnownJobID(jobName string, jobID int64) {
	w.Lock()
	defer w.Unlock()

	w.jobs[jobName] = jobID
}

func (w *watcher) Watch(jobName string, lastKnownJobID int64) {
	w.setLastKnownJobID(jobName, lastKnownJobID)
}

func (w *watcher) jobNames() []string {
	w.RLock()
	defer w.RUnlock()

	var names []string

	for name := range w.jobs {
		names = append(names, name)
	}

	return names
}

func (w *watcher) lastKnownJobID(jobName string) int64 {
	w.RLock()
	defer w.RUnlock()

	return w.jobs[jobName]
}

// NewWatcher returns a worker pool that notifies for new Prow jobs.
func NewWatcher(ctx context.Context, storageClient *storage.BucketHandle, ticker <-chan time.Time) *watcher {
	ch := make(chan job.Job)
	w := watcher{
		jobs: make(map[string]int64),
		Jobs: ch,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				log.Println("tick")
				var wg sync.WaitGroup
				for _, jobName := range w.jobNames() {
					wg.Add(1)
					ctx, cancel := context.WithCancel(ctx)
					defer cancel()
					lastKnownID := w.lastKnownJobID(jobName)
					go func(jobName string) {
						defer wg.Done()
						for j := range FindNext(ctx, storageClient, jobName, lastKnownID) {
							jobID, err := strconv.ParseInt(j.Status.BuildID, 10, 64)
							if err != nil {
								log.Printf("error parsing the job ID: %v", err)
								break
							}
							ch <- j
							w.setLastKnownJobID(jobName, jobID)
						}
					}(jobName)
				}
				wg.Wait()
			}
		}
	}()

	return &w
}
