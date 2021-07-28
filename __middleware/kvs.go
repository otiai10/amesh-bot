package middleware

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

// KVSer ...
type KVSer interface {
	Set(path string, value interface{}) error
	Get(path string, dest interface{}) error
	Close() error
}

// KVS ...
func KVS(ctx context.Context, project string) KVSer {
	if os.Getenv("GAE_APPLICATION") == "" {
		return &LocalKVS{}
	}
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
