package job

import (
	"context"
	"encoding/json"
	"io"

	"cloud.google.com/go/storage"
)

func getReadCloser(ctx context.Context, bkt *storage.BucketHandle, path string) (io.ReadCloser, error) {
	return bkt.Object(path).NewReader(ctx)
}

func getJSON(ctx context.Context, bkt *storage.BucketHandle, path string, t any) error {
	rc, err := getReadCloser(ctx, bkt, path)
	if err != nil {
		return err
	}
	defer func() {
		if e := rc.Close(); e != nil && err == nil {
			err = e
		}
	}()
	err = json.NewDecoder(rc).Decode(t)
	return err
}
