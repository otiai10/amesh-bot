package service

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
)

type Cloudstorage struct {
	BaseURL string
}

func (cs *Cloudstorage) Exists(ctx context.Context, bucket string, name string) (exists bool, err error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()
	attrs, err := client.Bucket(bucket).Object(name).Attrs(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return false, fmt.Errorf("failed to fetch attribute of an object: %v", err)
	}
	return attrs != nil && attrs.Size != 0, nil
}

func (cs *Cloudstorage) Upload(ctx context.Context, bucket string, name string, contents []byte) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()
	writer := client.Bucket(bucket).Object(name).NewWriter(ctx)
	if _, err = writer.Write(contents); err != nil {
		return fmt.Errorf("failed to upload image to cloud storage: %v", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to terminate cloud storage client: %v", err)
	}
	return nil
}

// URL ...
// return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, name)
func (cs *Cloudstorage) URL(bucket string, name string) string {
	return fmt.Sprintf("%s/%s/%s", cs.BaseURL, bucket, name)
}
