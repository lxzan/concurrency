package queues

import (
	"github.com/lxzan/concurrency/internal"
	"github.com/lxzan/concurrency/logs"
	"runtime"
	"time"
	"unsafe"
)

type Option func(o *options)

// WithConcurrency 设置最大并发
func WithConcurrency(n int64) Option {
	return func(o *options) {
		o.concurrency = n
	}
}

// WithTimeout 设置退出等待超时时间
func WithTimeout(t time.Duration) Option {
	return func(o *options) {
		o.timeout = t
	}
}

// WithLogger 设置日志组件
func WithLogger(logger logs.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithMultiple 设置多重队列, 降低锁竞争开销
// 注意: n会被转化为pow(2,x)
func WithMultiple(n int64) Option {
	return func(o *options) {
		o.multiple = true
		o.size = internal.ToBinaryNumber(n)
	}
}

// WithRecovery 设置恢复程序
func WithRecovery() Option {
	return func(o *options) {
		o.caller = func(logger logs.Logger, f func()) {
			defer func() {
				if e := recover(); e != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					msg := *(*string)(unsafe.Pointer(&buf))
					logger.Errorf("fatal error: %v\n%v\n", e, msg)
				}
			}()

			f()
		}
	}
}
