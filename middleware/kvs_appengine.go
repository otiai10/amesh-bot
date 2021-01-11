package middleware

import (
	"context"

	"cloud.google.com/go/firestore"
)

// KVSClient ...
type KVSClient struct {
	Project string
	Context context.Context
	*firestore.Client
}

// Close ...
func (kvs *KVSClient) Close() error {
	return kvs.Client.Close()
}

// Set ...
func (kvs *KVSClient) Set(path string, value interface{}) error {
	_, err := kvs.Client.Doc(path).Set(kvs.Context, value)
	return err
}

// Get ...
func (kvs *KVSClient) Get(path string, dest interface{}) error {
	doc, err := kvs.Client.Doc(path).Get(kvs.Context)
	if err != nil {
		return err
	}
	return doc.DataTo(dest)
}
