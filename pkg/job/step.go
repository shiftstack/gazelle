package job

import (
	"context"
	"io"
	"path"
	"time"

	"cloud.google.com/go/storage"
)

type Steps struct {
	storageClient       *storage.BucketHandle
	ciOperatorStepGraph CIOperatorStepGraph
}

// CIOperatorStepGraph is a copy of ci-tools' CIOperatorStepGraph
// https://github.com/openshift/ci-tools/blob/0d98dc14ccfb82a39dd38f0f99b5874ad104f2e4/pkg/api/graph.go#L518
type CIOperatorStepGraph []CIOperatorStepDetails

type CIOperatorStepDetails struct {
	CIOperatorStepDetailInfo `json:",inline"`
	Substeps                 []CIOperatorStepDetailInfo `json:"substeps,omitempty"`
}

type CIOperatorStepDetailInfo struct {
	StepName     string         `json:"name"`
	Description  string         `json:"description"`
	Dependencies []string       `json:"dependencies"`
	StartedAt    *time.Time     `json:"started_at"`
	FinishedAt   *time.Time     `json:"finished_at"`
	Duration     *time.Duration `json:"duration,omitempty"`
	LogURL       string         `json:"log_url,omitempty"`
	Failed       *bool          `json:"failed,omitempty"`
	// Manifests    []ctrlruntimeclient.Object `json:"manifests,omitempty"`
}

type Step struct {
	path string

	Finished
}

func (s Step) Name() string {
	return path.Base(s.path)
}

func (s Step) BuildLog(ctx context.Context, storageClient *storage.BucketHandle) (io.ReadCloser, error) {
	return getReadCloser(ctx, storageClient, path.Join(s.path, "build-log.txt"))
}

func (s Step) GetArtifact(ctx context.Context, storageClient *storage.BucketHandle, name string) (io.ReadCloser, error) {
	return storageClient.Object(path.Join(s.path, "artifacts", name)).NewReader(ctx)
}
