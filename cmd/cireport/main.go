package main

import (
	"context"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"cloud.google.com/go/storage"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/prow"
	"github.com/shiftstack/gazelle/pkg/rca"
	"github.com/shiftstack/gazelle/pkg/slack"
	"google.golang.org/api/option"

	prowjobv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const debug = false

var tmpl = template.Must(template.New("notify").Parse("Job *{{.Spec.Job}}* ended with *{{.Status.State}}*. <{{.Status.URL}}|View logs>{{ range .Reports }}\n * *{{ .Name }}*: {{ . }}{{ end }}"))

func getClient(ctx context.Context) (*storage.BucketHandle, error) {
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}

	return client.Bucket("origin-ci-test"), nil
}

func main() {
	slackHook := os.Getenv("SLACK_HOOK")

	ctx := context.Background()

	reportJobsFrom := time.Now()
	ticker := time.NewTicker(time.Hour)

	if debug {
		reportJobsFrom = time.Now().Add(-7 * 24 * time.Hour)
		ticker.Reset(time.Second)
		time.Sleep(1200 * time.Millisecond)
		ticker.Stop()
	}

	defer ticker.Stop()

	notifier := slack.New(slackHook)

	storageClient, err := getClient(ctx)
	if err != nil {
		panic(err)
	}

	watcher := prow.NewWatcher(ctx, storageClient, ticker.C)

	for _, jobName := range jobsToBeWatched {
		log.Printf("Adding %s", jobName)
		watcher.Watch(jobName, 0)
	}

	for j := range watcher.Jobs {
		if j.Status.CompletionTime.After(reportJobsFrom) {
			switch j.Status.State {
			case prowjobv1.FailureState, prowjobv1.ErrorState, prowjobv1.AbortedState:
				var reports []rca.Report
				for _, check := range [...]func(context.Context, *storage.BucketHandle, job.Job) (rca.Report, error){
					rca.ServerError,
					rca.PreviousResult,
				} {
					report, err := check(ctx, storageClient, j)
					if err != nil {
						log.Printf("rca error: %v", err)
					} else if report.Occurred() {
						reports = append(reports, report)
					}
				}

				var s strings.Builder
				if err := tmpl.ExecuteTemplate(&s, "notify", struct {
					job.Job
					Reports []rca.Report
				}{
					Job:     j,
					Reports: reports,
				}); err != nil {
					log.Printf("ERROR: failed to build the notification message: %v", err)
					continue
				}

				if err := notifier.Send(ctx, s.String()); err != nil {
					log.Printf("failed to send notification: %v", err)
				}
			}
		}
	}
}
