package concurrency

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
)

const DefaultConcurrency = 8

var (
	// 默认日志组件
	DefaultLogger = new(loggerWrapper)

	// 默认任务调用器
	DefaultCaller CallerFunc = func(job Job) error {
		return job.Do(job.Args)
	}

	// 默认工作队列, 64协程, 开启恢复模式
	DefaultWorkerQueue = NewWorkerQueue(
		WithConcurrency(64),
		WithRecovery(),
	)
)

func init() {
	DefaultWorkerQueue.OnError = func(err error) {
		log.Printf(err.Error())
	}
}

type (
	Job struct {
		Args interface{}
		Do   func(args interface{}) error
	}

	Config struct {
		Context     context.Context // 上下文
		Concurrency int64           // 并发协程数量
		Caller      CallerFunc      // 任务调用器
		LogEnabled  bool            // 是否开启日志, 默认开启
		Logger      Logger          // 日志组件
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
	if c.Logger == nil {
		c.Logger = DefaultLogger
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

// 设置是否开启日志
func WithLogEnabled(enabled bool) Option {
	return func(c *Config) {
		c.LogEnabled = enabled
	}
}

// 设置日志组件
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithRecovery 设置恢复模式
func WithRecovery() Option {
	return func(c *Config) {
		c.Caller = func(job Job) (err error) {
			defer func() {
				if fatalError := recover(); fatalError != nil {
					var msg = make([]byte, 0, 256)
					msg = append(msg, fmt.Sprintf("fatal error: %v\n", fatalError)...)
					for i := 1; true; i++ {
						_, caller, line, ok := runtime.Caller(i)
						if !ok {
							break
						}
						if !strings.Contains(caller, "src/runtime") {
							msg = append(msg, fmt.Sprintf("caller: %s, line: %d\n", caller, line)...)
						}
					}
					err = errors.New(string(msg))
				}
			}()
			return job.Do(job.Args)
		}
	}
}
