package concurrency

import (
	"context"
	"errors"
	"fmt"
	"log"
)

const (
	DefaultConcurrency = 8
)

var (
	DefaultLogger = new(loggerWrapper)

	DefaultCaller CallerFunc = func(job Job) error {
		return job.Do(job.Args)
	}
)

type loggerWrapper struct{}

func (c *loggerWrapper) Errorf(format string, v ...interface{}) {
	log.Default().Printf(format, v...)
}

type (
	Logger interface {
		Errorf(format string, v ...interface{})
	}

	Config struct {
		err         error
		Context     context.Context
		Concurrency int64
		Caller      CallerFunc
		Logger      Logger
	}

	Option func(c *Config)

	CallerFunc func(job Job) error
)

func (c *Config) init() *Config {
	if c.Context == nil {
		c.Context = context.Background()
	}
	if c.Concurrency <= 0 {
		c.Concurrency = DefaultConcurrency
	}
	if c.Logger == nil {
		c.Logger = DefaultLogger
	}
	if c.Caller == nil {
		c.Caller = DefaultCaller
	}
	return c
}

func WithConcurrency(x int64) Option {
	return func(c *Config) {
		c.Concurrency = x
	}
}

func WithContext(ctx context.Context) Option {
	return func(c *Config) {
		c.Context = ctx
	}
}

func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithRecovery() Option {
	return func(c *Config) {
		c.Caller = func(job Job) error {
			defer func() {
				if err := recover(); err != nil {
					var msg = fmt.Sprintf("fatal error: %v", err)
					c.err = errors.New(msg)
					c.Logger.Errorf(msg)
				}
			}()
			return job.Do(job.Args)
		}
	}
}
