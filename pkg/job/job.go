package job

import (
	"context"
	"fmt"
	"io"
	"path"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	prowjobv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

type Job struct {
	path string
	prowjobv1.ProwJob

	stepGraph      CIOperatorStepGraph
	errNoStepgraph error
}

func New(ctx context.Context, storageClient *storage.BucketHandle, jobPath string) (Job, error) {
	var prowJob prowjobv1.ProwJob
	if err := getJSON(ctx, storageClient, path.Join(jobPath, "prowjob.json"), &prowJob); err != nil {
		return Job{}, fmt.Errorf("failed to get ProwJob: %w", err)
	}

	return NewWithProwJob(ctx, storageClient, jobPath, prowJob)
}

func NewWithProwJob(ctx context.Context, storageClient *storage.BucketHandle, jobPath string, prowJob prowjobv1.ProwJob) (Job, error) {
	var (
		stepGraph      CIOperatorStepGraph
		errNoStepgraph error
	)
	if err := getJSON(ctx, storageClient, path.Join(jobPath, "artifacts", "ci-operator-step-graph.json"), &stepGraph); err != nil {
		errNoStepgraph = NewErrNoStepgraph(err)
	}

	return Job{
		path:           jobPath,
		ProwJob:        prowJob,
		stepGraph:      stepGraph,
		errNoStepgraph: errNoStepgraph,
	}, nil
}

func (j Job) BuildLog(ctx context.Context, storageClient *storage.BucketHandle) (io.ReadCloser, error) {
	return storageClient.Object(path.Join(j.path, "build-log.txt")).NewReader(ctx)
}

func (j Job) stepsPath() string {
	if j.errNoStepgraph != nil {
		return ""
	}
	multistageStep := j.stepGraph[len(j.stepGraph)-1]
	return path.Join(j.path, "artifacts", multistageStep.StepName) + "/"
}

func (j Job) StepNames(ctx context.Context, storageClient *storage.BucketHandle) ([]string, error) {
	if j.errNoStepgraph != nil {
		return nil, j.errNoStepgraph
	}

	var steps []string
	query := &storage.Query{
		Prefix:    j.stepsPath(),
		Delimiter: "/",
	}
	query.SetAttrSelection([]string{"Prefix"})
	objects := storageClient.Objects(ctx, query)
	for {
		attrs, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		steps = append(steps, path.Base(attrs.Prefix))
	}

	return steps, nil
}

func (j Job) Step(ctx context.Context, storageClient *storage.BucketHandle, stepName string) (Step, error) {
	if j.errNoStepgraph != nil {
		return Step{}, j.errNoStepgraph
	}

	stepPath := path.Join(j.stepsPath(), stepName)

	var finished Finished
	if err := getJSON(ctx, storageClient, path.Join(stepPath, "finished.json"), &finished); err != nil {
		return Step{}, fmt.Errorf("failed to get finished.json: %v", err)
	}

	return Step{
		path:     stepPath,
		Finished: finished,
	}, nil
}
