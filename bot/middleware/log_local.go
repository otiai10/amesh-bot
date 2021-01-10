//+build !appengine

package middleware

import (
	"context"
	"log"
)

// Log ...
func Log(ctx context.Context, project string, name string) Logger {
	return &LogClient{
		Project: project,
		Name:    name,
	}
}

// Debug ...
func (c *LogClient) Debug(entry interface{}, labels Labels) {
	log.Printf("[%s]\t%v\t%+v\n", "DEBUG", labels, entry)
}

// Info ...
func (c *LogClient) Info(entry interface{}, labels Labels) {
	log.Printf("[%s]\t%v\t%+v\n", "INFO", labels, entry)
}

// Error ...
func (c *LogClient) Error(entry interface{}, labels Labels) {
	log.Printf("[%s]\t%v\t%+v\n", "ERROR", labels, entry)
}

// Critical ...
func (c *LogClient) Critical(entry interface{}, labels Labels) {
	log.Fatalf("[%s]\t%v\t%+v\n", "CRITICAL", labels, entry)
}

// Close ...
func (c *LogClient) Close() error {
	return nil
}
