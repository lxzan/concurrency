package logs

import "log"

var DefaultLogger = new(logger)

type Logger interface {
	Errorf(format string, args ...any)
}

type logger struct{}

func (c *logger) Errorf(format string, args ...any) {
	log.Printf(format, args...)
}
