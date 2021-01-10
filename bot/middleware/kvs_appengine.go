//+build appengine

package middleware

import (
	"context"

	"cloud.google.com/go/firestore"
)

// KVS ...
func KVS(ctx context.Context, project string) *KVSClient {
	client, err := firestore.NewClient(ctx, project)
	return &KVSClient{
		Project: project,
		Client:  client,
		Context: ctx,
		Err:     err,
	}
}

func (kvs *KVSClient) Close() error {
	if kvs.Err != nil {
		return kvs.Err
	}
	return kvs.Client.Close()
}

func (kvs *KVSClient) Set(path string, value interface{}) error {
	if kvs.Err != nil {
		return kvs.Err
	}
	_, err := client.Doc(path).Set(kvs.Context, value)
	kvs.Err = err
	return err
}

func (kvs *KVSClient) Get(path string, dest interface{}) error {
	if kvs.Err != nil {
		return kvs.Err
	}
	doc, err := client.Doc(path).Get(kvs.Context)
	if err != nil {
		kvs.Err = err
		return kvs.Err
	}
	err := doc.DataTo(dest)
	kvs.Err = err
	return kvs.Err
}
