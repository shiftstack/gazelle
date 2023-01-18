package rca

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/prow"
)

type Report struct {
	name     string
	occurred bool
	message  string
}

func (r Report) Name() string {
	return r.name
}

func (r Report) Occurred() bool {
	return r.occurred
}
func (r Report) String() string {
	return r.message
}

// ServerError detects failed VMs
func ServerError(ctx context.Context, storageClient *storage.BucketHandle, j job.Job) (Report, error) {
	step, err := j.Step(ctx, storageClient, "openstack-gather")
	if err != nil {
		return Report{}, fmt.Errorf("failed to get openstack-gather: %w", err)
	}

	if !step.Passed {
		return Report{}, nil
	}

	rc, err := step.GetArtifact(ctx, storageClient, "json/openstack_server_list.json")
	if err != nil {
		return Report{}, fmt.Errorf("failed to get openstack_server_list.json: %w", err)
	}
	defer rc.Close()

	var servers []struct{ Status string }
	if err := json.NewDecoder(rc).Decode(&servers); err != nil {
		return Report{}, fmt.Errorf("failed to decode openstack_server_list.json: %w", err)
	}

	for _, server := range servers {
		if server.Status != "ACTIVE" {
			return Report{
				name:     "ServerError",
				occurred: true,
				message:  "one or more servers are not ACTIVE",
			}, nil
		}
	}

	return Report{}, nil
}

// PreviousResult checks whether the previous run of the same job succeeded
func PreviousResult(ctx context.Context, storageClient *storage.BucketHandle, j job.Job) (Report, error) {
	previousJob, ok := <-prow.FindPreviousN(ctx, storageClient, j.Spec.Job, j.Status.BuildID, 1)
	if !ok {
		return Report{}, nil
	}

	return Report{
		name:     "PreviousResult",
		occurred: true,
		message:  fmt.Sprintf("the previous run (%s) ended with %s", previousJob.Status.BuildID, previousJob.Status.State),
	}, nil
}
