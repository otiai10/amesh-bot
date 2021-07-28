package middleware

import (
	"cloud.google.com/go/logging"
)

// LogClient ...
type LogClient struct {
	Project string
	Name    string
	*logging.Client
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
func (c *LogClient) Close() error {
	return c.Client.Close()
}
