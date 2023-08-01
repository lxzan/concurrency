package queues

import (
	"log"
	"runtime"
	"unsafe"
)

type Option func(o *options)

// WithConcurrency 设置最大并发
func WithConcurrency(num int64) Option {
	return func(o *options) {
		o.concurrency = num
	}
}

// WithRecovery 设置恢复程序
func WithRecovery() Option {
	return func(o *options) {
		o.caller = func(f func()) {
			defer func() {
				if e := recover(); e != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					msg := *(*string)(unsafe.Pointer(&buf))
					log.Printf("fatal error: %v\n%v\n", e, msg)
				}
			}()

			f()
		}
	}
}
