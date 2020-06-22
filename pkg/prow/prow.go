package prow

import (
	"context"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const bucketName = "origin-ci-test"

type int64slice []int64

func (s int64slice) Len() int           { return len([]int64(s)) }
func (s int64slice) Less(i, j int) bool { return s[i] < s[j] }
func (s int64slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

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

func JobIDs(ctx context.Context, jobName string, from int64) <-chan int64 {
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		panic(err)
	}

	bkt := client.Bucket(bucketName)

	lastJobID, err := getLastID(ctx, bkt, jobName)
	if err != nil {
		panic(err)
	}

	ids := make(chan int64)
	if from > lastJobID {
		close(ids)
		return ids
	}

	query := &storage.Query{
		Prefix:    "logs/" + jobName + "/",
		Delimiter: "/",
	}
	query.SetAttrSelection([]string{"Prefix"})

	go func() {
		defer close(ids)
		objects := bkt.Objects(ctx, query)
		for {
			select {
			case <-ctx.Done():
				panic(ctx.Err())
			default:
			}

			attrs, err := objects.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				panic(err)
			}

			path := strings.Split(attrs.Prefix, "/")
			if len(path) > 2 {
				id, err := strconv.ParseInt(path[len(path)-2], 10, 64)
				if err != nil {
					panic(err)
				}
				if id < from {
					continue
				}
				if id == lastJobID {
					finished, err := jobIsFinished(ctx, bkt, jobName, id)
					if err != nil {
						panic(err)
					}
					if !finished {
						continue
					}
				}
				ids <- id
			}
		}
	}()
	return ids
}

func Sorted(unsorted <-chan int64) <-chan int64 {
	var cache []int64
	for n := range unsorted {
		cache = append(cache, n)
	}

	sort.Sort(int64slice(cache))

	sorted := make(chan int64)
	go func() {
		defer close(sorted)
		for _, n := range cache {
			sorted <- n
		}
	}()

	return sorted
}
