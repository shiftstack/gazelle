package prow

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/shiftstack/gazelle/pkg/job"
	"google.golang.org/api/iterator"
	prowjobv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const bucketName = "origin-ci-test"

func jobIsFinished(ctx context.Context, bkt *storage.BucketHandle, jobName string, jobID int64) (bool, error) {
	_, err := bkt.Object("logs/" + jobName + "/" + strconv.FormatInt(jobID, 10) + "/finished.json").Attrs(ctx)
	switch err {
	case nil:
		return true, nil
	case storage.ErrObjectNotExist:
		return false, nil
	default:
		return false, err
	}
}

func getProwJob(ctx context.Context, bkt *storage.BucketHandle, jobName string, jobID int64) (prowjobv1.ProwJob, error) {
	var pj prowjobv1.ProwJob
	r, err := bkt.Object("logs/" + jobName + "/" + strconv.FormatInt(jobID, 10) + "/prowjob.json").NewReader(ctx)
	if err != nil {
		return pj, err
	}
	defer r.Close()

	err = json.NewDecoder(r).Decode(&pj)
	return pj, err
}

func getLastID(ctx context.Context, bkt *storage.BucketHandle, jobName string) (int64, error) {
	obj := bkt.Object("logs/" + jobName + "/latest-build.txt")

	r, err := obj.NewReader(ctx)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	s, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(string(s), 10, 64)
}

func FindNext(ctx context.Context, storageClient *storage.BucketHandle, jobName string, lastKnownID int64) <-chan job.Job {
	lastJobID, err := getLastID(ctx, storageClient, jobName)
	if err != nil {
		panic("failed to get the last ID for " + jobName + " :" + err.Error())
	}

	jobCh := make(chan job.Job)
	if lastKnownID > lastJobID {
		close(jobCh)
		return jobCh
	}

	query := &storage.Query{
		Prefix:    "logs/" + jobName + "/",
		Delimiter: "/",
	}
	query.SetAttrSelection([]string{"Prefix"})

	go func() {
		defer close(jobCh)
		objects := storageClient.Objects(ctx, query)

		for {
			attrs, err := objects.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				log.Print(err)
				return
			}

			jobPath := strings.Split(attrs.Prefix, "/")
			if len(jobPath) > 2 {
				id, err := strconv.ParseInt(jobPath[len(jobPath)-2], 10, 64)
				if err != nil {
					log.Print(err)
					return
				}
				if id <= lastKnownID {
					return
				}
				if id > lastJobID {
					finished, err := jobIsFinished(ctx, storageClient, jobName, id)
					if err != nil {
						log.Print(err)
						return
					}
					if !finished {
						return
					}
				}

				j, err := job.New(ctx, storageClient, path.Join("logs", jobName, strconv.FormatInt(id, 10)))
				if err != nil {
					log.Print(err)
					break
				}

				select {
				case jobCh <- j:
				case <-ctx.Done():
					log.Print(err)
					break
				}
			}
		}
	}()
	return jobCh
}

// min returns the smallest between a and n, or a if n is negative.
func min(a, n int) int {
	if n < 0 || a < n {
		return a
	}
	return n
}

func FindPrevious(ctx context.Context, storageClient *storage.BucketHandle, jobName string, currentID string) <-chan job.Job {
	return FindPreviousN(ctx, storageClient, jobName, currentID, -1)
}

func FindPreviousN(ctx context.Context, storageClient *storage.BucketHandle, jobName string, currentID string, n int) <-chan job.Job {
	currentJobID, err := strconv.ParseInt(currentID, 10, 64)
	if err != nil {
		panic("failed to parse the job ID as int64: " + err.Error())
	}

	jobCh := make(chan job.Job)

	query := &storage.Query{
		Prefix:    "logs/" + jobName + "/",
		Delimiter: "/",
	}
	query.SetAttrSelection([]string{"Prefix"})
	objects := storageClient.Objects(ctx, query)

	var objectList []*storage.ObjectAttrs
	for {
		attrs, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Print(err)
		}
		objectList = append(objectList, attrs)
	}
	reversedObjectList := make([]*storage.ObjectAttrs, min(len(objectList), n))

	for i := range reversedObjectList {
		reversedObjectList[i] = objectList[len(objectList)-1-1-i]
	}

	go func() {
		defer close(jobCh)

		for _, attrs := range reversedObjectList {
			jobPath := strings.Split(attrs.Prefix, "/")
			if len(jobPath) > 2 {
				id, err := strconv.ParseInt(jobPath[len(jobPath)-2], 10, 64)
				if err != nil {
					log.Print(err)
					break
				}
				if id >= currentJobID {
					continue
				}

				j, err := job.New(ctx, storageClient, path.Join("logs", jobName, strconv.FormatInt(id, 10)))
				if err != nil {
					log.Print(err)
					break
				}

				select {
				case jobCh <- j:
				case <-ctx.Done():
					log.Print(err)
					break
				}
			}
		}
	}()
	return jobCh
}
