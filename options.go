package concurrency

import (
	"context"
	"errors"
	"fmt"
	"runtime"
)

const (
	DefaultConcurrency = 8
)

var (
	DefaultCaller CallerFunc = func(job Job) error {
		return job.Do(job.Args)
	}
)

type (
	Config struct {
		Context     context.Context
		Concurrency int64
		Caller      CallerFunc
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

func WithRecovery() Option {
	return func(c *Config) {
		c.Caller = func(job Job) (err error) {
			defer func() {
				if e := recover(); e != nil {
					_, caller, line, _ := runtime.Caller(2)
					var msg = fmt.Sprintf("fatal error: %v, caller: %s, line: %d", e, caller, line)
					err = errors.New(msg)
				}
			}()
			return job.Do(job.Args)
		}
	}
}
