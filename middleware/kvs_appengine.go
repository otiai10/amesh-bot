package middleware

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

// KVS ...
func KVS(ctx context.Context, project string) *KVSClient {
	client, err := firestore.NewClient(ctx, project)
	if err != nil {
		log.Fatalln("[Firestore Client]", err)
	}
	return &KVSClient{
		Project: project,
		Client:  client,
		Context: ctx,
	}
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
