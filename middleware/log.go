package middleware

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/logging"
)

// Log ...
func Log(ctx context.Context, project string, name string) Logger {
	if os.Getenv("GAE_APPLICATION") == "" {
		return &LocalLogger{}
	}
	client, err := logging.NewClient(ctx, project)
	if err != nil {
		log.Fatalln("[Cloud Logging]", err)
	}
	return &LogClient{Project: project, Name: name, Client: client}
}

// Labels ...
type Labels map[string]string

// String ...
func (labels Labels) String() string {
	var elems []string
	for key, val := range labels {
		elems = append(elems, fmt.Sprintf("%s=%s", key, val))
	}
	return strings.Join(elems, "\t")
}

// Logger ...
type Logger interface {
	Debug(entry interface{}, labels Labels)
	Info(entry interface{}, labels Labels)
	Error(entry interface{}, labels Labels)
	Critical(entry interface{}, labels Labels)
	Close() error
}
