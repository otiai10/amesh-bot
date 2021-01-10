package middleware

import (
	"context"

	"cloud.google.com/go/firestore"
)

// KVSClient ...
type KVSClient struct {
	Project string
	Err     error
	Context context.Context
	*firestore.Client
}
