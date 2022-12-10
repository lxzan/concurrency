package concurrency

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
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

// WithConcurrency 设置并发数量
func WithConcurrency(x int64) Option {
	return func(c *Config) {
		c.Concurrency = x
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(c *Config) {
		c.Context = ctx
	}
}

// WithRecovery 设置恢复模式
func WithRecovery() Option {
	return func(c *Config) {
		c.Caller = func(job Job) (err error) {
			defer func() {
				if fatalError := recover(); fatalError != nil {
					var msg = make([]byte, 0, 128)
					msg = append(msg, fmt.Sprintf("fatal error: %v\n", fatalError)...)
					for i := 1; true; i++ {
						if _, caller, line, ok := runtime.Caller(i); ok {
							if !strings.Contains(caller, "src/runtime") {
								msg = append(msg, fmt.Sprintf("caller: %s, line: %d\n", caller, line)...)
							}
						} else {
							break
						}
					}
					err = errors.New(string(msg))
				}
			}()
			return job.Do(job.Args)
		}
	}
}
