package middleware

import (
	"os"
	"reflect"
	"strings"
)

// import "context"

// LocalKVS ...
type LocalKVS struct{}

// Close ...
func (kvs *LocalKVS) Close() error {
	return nil
}

// Set ...
func (kvs *LocalKVS) Set(path string, value interface{}) error {
	return nil
}

// Get ...
func (kvs *LocalKVS) Get(path string, dest interface{}) error {
	if strings.HasPrefix(path, "Teams/") {
		reflect.ValueOf(dest).Elem().
			FieldByName("AccessToken").
			Set(reflect.ValueOf(os.Getenv("SLACK_BOT_USER_OAUTH_ACCESS_TOKEN")))
	}
	return nil
}
