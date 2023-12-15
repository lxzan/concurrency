package queues

import (
	"context"
	"github.com/lxzan/concurrency/logs"
	"time"
)

var defaultCaller Caller = func(logger logs.Logger, f func()) { f() }

const (
	defaultSharding    = 1
	defaultConcurrency = 8
	defaultTimeout     = 30 * time.Second
)

type (
	Caller func(logger logs.Logger, f func())

	Job func()

	Queue interface {
		// Len 获取队列中剩余任务数量
		Len() int

		// Push 追加任务
		Push(job Job)

		// Stop 停止
		// 停止后不能追加新的任务, 队列中剩余的任务会继续执行, 到收到上下文信号为止.
		Stop(ctx context.Context) error
	}
)

func New(opts ...Option) Queue {
	opts = append(opts, withInitialize())
	o := new(options)
	for _, f := range opts {
		f(o)
	}

	if o.sharding == 1 {
		return newSingleQueue(o)
	}
	return newMultipleQueue(o)
}
