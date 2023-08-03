package queues

import (
	"github.com/lxzan/concurrency/logs"
	"github.com/pkg/errors"
	"time"
)

const (
	defaultConcurrency = 8
	defaultTimeout     = 30 * time.Second
)

var ErrStopTimeout = errors.New("stop timeout")

type (
	options struct {
		multiple    bool
		size        int64
		concurrency int64
		timeout     time.Duration
		caller      Caller
		logger      logs.Logger
	}

	Caller func(logger logs.Logger, f func())

	Job func()

	Queue interface {
		// 追加任务
		Push(job Job)

		// 停止
		// 注意: 此方法有阻塞等待任务结束逻辑; 停止后调用Push方法不会产生任何效果.
		Stop() error
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
