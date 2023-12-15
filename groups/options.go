package groups

import (
	"github.com/lxzan/concurrency/internal"
	"github.com/pkg/errors"
	"runtime"
	"time"
	"unsafe"
)

type options struct {
	timeout     time.Duration
	concurrency int64
	caller      Caller
}

type Option func(o *options)

// WithTimeout 设置任务超时时间
func WithTimeout(t time.Duration) Option {
	return func(o *options) {
		o.timeout = t
	}
}

// WithConcurrency 设置最大并发
func WithConcurrency(n uint32) Option {
	return func(o *options) {
		o.concurrency = int64(n)
	}
}

// WithRecovery 设置恢复程序
func WithRecovery() Option {
	return func(o *options) {
		o.caller = func(args any, f func(any) error) (err error) {
			defer func() {
				if e := recover(); e != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					msg := *(*string)(unsafe.Pointer(&buf))
					err = errors.New(msg)
				}
			}()

			return f(args)
		}
	}
}

func withInitialize() Option {
	return func(o *options) {
		o.timeout = internal.SelectValue(o.timeout <= 0, defaultWaitTimeout, o.timeout)
		o.concurrency = internal.SelectValue(o.concurrency <= 0, defaultConcurrency, o.concurrency)
		o.caller = internal.SelectValue(o.caller == nil, defaultCaller, o.caller)
	}
}
