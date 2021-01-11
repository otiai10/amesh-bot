package middleware

import "log"

// LocalLogger ...
type LocalLogger struct{}

// Debug ...
func (c *LocalLogger) Debug(entry interface{}, labels Labels) {
	log.Printf("[%s]\t%v\t%+v\n", "DEBUG", labels, entry)
}

// Info ...
func (c *LocalLogger) Info(entry interface{}, labels Labels) {
	log.Printf("[%s]\t%v\t%+v\n", "INFO", labels, entry)
}

// Error ...
func (c *LocalLogger) Error(entry interface{}, labels Labels) {
	log.Printf("[%s]\t%v\t%+v\n", "ERROR", labels, entry)
}

// Critical ...
func (c *LocalLogger) Critical(entry interface{}, labels Labels) {
	log.Fatalf("[%s]\t%v\t%+v\n", "CRITICAL", labels, entry)
}

// Close ...
func (c *LocalLogger) Close() error {
	return nil
}
