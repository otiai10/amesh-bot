//+build appengine

package middleware

import (
	"context"
	"log"

	"cloud.google.com/go/logging"
)

// Log ...
func Log(ctx context.Context, project string, name string) Logger {
	client, err := logging.NewClient(ctx, project)
	if err != nil {
		log.Fatalln("[Cloud Logging]", err)
	}
	return &LogClient{Project: project, Name: name, Client: client}
}

// Debug ...
func (c *LogClient) Debug(entry interface{}, labels Labels) {
	c.Client.Logger(c.Name).Log(logging.Entry{
		Payload:  entry,
		Severity: logging.Debug,
		Labels:   labels,
	})
}

// Info ...
func (c *LogClient) Info(entry interface{}, labels Labels) {
	c.Client.Logger(c.Name).Log(logging.Entry{
		Payload:  entry,
		Severity: logging.Info,
		Labels:   labels,
	})
}

// Error ...
func (c *LogClient) Error(entry interface{}, labels Labels) {
	c.Client.Logger(c.Name).Log(logging.Entry{
		Payload:  entry,
		Severity: logging.Error,
		Labels:   labels,
	})
}

// Critical ...
func (c *LogClient) Critical(entry interface{}, labels Labels) {
	c.Client.Logger(c.Name).Log(logging.Entry{
		Payload:  entry,
		Severity: logging.Critical,
		Labels:   labels,
	})
}

// Close ...
func (c *LogClient) Close() {
	return c.Client.Close()
}
