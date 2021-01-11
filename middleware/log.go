package middleware

import (
	"fmt"
	"strings"

	"cloud.google.com/go/logging"
)

// LogClient ...
type LogClient struct {
	Project string
	Name    string
	*logging.Client
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
