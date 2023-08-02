package groups

import (
	"github.com/pkg/errors"
	"runtime"
	"time"
	"unsafe"
)

type Option func(o *options)

// WithTimeout 设置任务超时时间
func WithTimeout(t time.Duration) Option {
	return func(o *options) {
		o.timeout = t
	}
}

// WithConcurrency 设置最大并发
func WithConcurrency(num int64) Option {
	return func(o *options) {
		o.concurrency = num
	}
}

// WithCancel 设置遇到错误放弃执行剩余任务
func WithCancel() Option {
	return func(o *options) {
		o.cancel = true
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
