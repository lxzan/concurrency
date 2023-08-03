package queues

import (
	"github.com/lxzan/concurrency/logs"
	"time"
)

const (
	defaultConcurrency = 8
	defaultTimeout     = 30 * time.Second
)

type (
	options struct {
		multiple    bool
		concurrency int64
		timeout     time.Duration
		caller      Caller
		logger      logs.Logger
	}

	Caller func(logger logs.Logger, f func())

	Job func()

	Queue interface {
		Push(job Job)
		Stop()
	}
)

func New(opts ...Option) Queue {
	o := &options{
		multiple:    false,
		concurrency: defaultConcurrency,
		timeout:     defaultTimeout,
		caller:      func(logger logs.Logger, f func()) { f() },
		logger:      logs.DefaultLogger,
	}
	for _, f := range opts {
		f(o)
	}

	if o.multiple {
		return newMultipleQueue(o)
	}
	return newSingleQueue(o)
}
