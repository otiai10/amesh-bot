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
