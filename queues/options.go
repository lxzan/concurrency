package queues

import (
	"github.com/lxzan/concurrency/internal"
	"github.com/lxzan/concurrency/logs"
	"runtime"
	"time"
	"unsafe"
)

type options struct {
	sharding    int64         // 分片数
	concurrency uint32        // 并行度
	timeout     time.Duration // 退出等待超时时间
	caller      Caller        // 调用器
	logger      logs.Logger   // 日志组件
}

type Option func(o *options)

// WithSharding 设置分片数量, 有利于降低锁竞争开销, 默认为1
func WithSharding(num uint32) Option {
	return func(o *options) {
		o.sharding = int64(num)
	}
}

// WithConcurrency 设置最大并行度, 默认为8
func WithConcurrency(n uint32) Option {
	return func(o *options) {
		o.concurrency = n
	}
}

// WithTimeout 设置退出等待超时时间, 默认30s
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

func withInitialize() Option {
	return func(o *options) {
		o.sharding = internal.SelectValue(o.sharding <= 0, defaultSharding, o.sharding)
		o.sharding = internal.ToBinaryNumber(o.sharding)
		o.concurrency = internal.SelectValue(o.concurrency <= 0, defaultConcurrency, o.concurrency)
		o.timeout = internal.SelectValue(o.timeout <= 0, defaultTimeout, o.timeout)
		o.logger = internal.SelectValue[logs.Logger](o.logger == nil, logs.DefaultLogger, o.logger)
		o.caller = internal.SelectValue(o.caller == nil, defaultCaller, o.caller)
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
