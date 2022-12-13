package concurrency

import "log"

type Logger interface {
	Errorf(layout string, v ...interface{})
}

type loggerWrapper struct{}

func (c *loggerWrapper) Errorf(layout string, v ...interface{}) {
	log.Default().Printf(layout, v...)
}
