package groups

import (
	"context"
	"errors"
	"github.com/lxzan/concurrency/internal"
	"sync"
	"time"
)

const (
	defaultConcurrency = 8                // 默认并发度
	defaultWaitTimeout = 60 * time.Second // 默认线程同步等待超时
)

type (
	Caller func(args any, f func(any) error) error

	Group[T any] struct {
		options   *options
		mu        *sync.Mutex        // 锁
		errs      []error            // 错误
		done      chan bool          // 信号
		q         []T                // 任务队列
		taskDone  int64              // 已完成任务数量
		taskTotal int64              // 总任务数量
		OnMessage func(args T) error // 任务处理
		OnError   func(err error)    // 错误处理
	}
)

// New 新建一个任务集
func New[T any](opts ...Option) *Group[T] {
	o := &options{
		timeout:     defaultWaitTimeout,
		concurrency: defaultConcurrency,
		caller:      func(args any, f func(any) error) error { return f(args) },
	}
	for _, f := range opts {
		f(o)
	}

	c := &Group[T]{
		options:  o,
		mu:       &sync.Mutex{},
		q:        make([]T, 0),
		taskDone: 0,
		done:     make(chan bool),
	}
	c.OnMessage = func(args T) error {
		return nil
	}
	c.OnError = func(err error) {}
	return c
}

func (c *Group[T]) clear() {
	c.mu.Lock()
	c.q = c.q[:0]
	c.mu.Unlock()
}

func (c *Group[T]) getJob() (v T, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n := len(c.q); n == 0 {
		return
	}
	var result = c.q[0]
	c.q = c.q[1:]
	return result, true
}

// incrAndIsDone
// 已完成任务+1, 并检查任务是否全部完成
func (c *Group[T]) incrAndIsDone() bool {
	c.mu.Lock()
	c.taskDone++
	ok := c.taskDone == c.taskTotal
	c.mu.Unlock()
	return ok
}

func (c *Group[T]) hasError() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.errs) > 0
}

func (c *Group[T]) do(args T) {
	if err := c.options.caller(args, func(v any) error {
		if c.options.cancel && c.hasError() {
			return nil
		}
		return c.OnMessage(v.(T))
	}); err != nil {
		c.mu.Lock()
		c.errs = append(c.errs, err)
		c.mu.Unlock()
		c.OnError(err)
	}

	if c.incrAndIsDone() {
		c.done <- true
		return
	}

	if nextJob, ok := c.getJob(); ok {
		c.do(nextJob)
	}
}

// Len 获取队列中剩余任务数量
func (c *Group[T]) Len() int {
	c.mu.Lock()
	x := len(c.q)
	c.mu.Unlock()
	return x
}

// Push 往任务队列中追加任务
func (c *Group[T]) Push(eles ...T) {
	c.mu.Lock()
	c.taskTotal += int64(len(eles))
	c.q = append(c.q, eles...)
	c.mu.Unlock()
}

// Update 线程安全操作
func (c *Group[T]) Update(f func()) {
	c.mu.Lock()
	f()
	c.mu.Unlock()
}

// Start 启动并等待所有任务执行完成
func (c *Group[T]) Start() error {
	var taskTotal = int64(c.Len())
	if taskTotal == 0 {
		return nil
	}

	var co = internal.Min(c.options.concurrency, taskTotal)
	for i := int64(0); i < co; i++ {
		if item, ok := c.getJob(); ok {
			go c.do(item)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.options.timeout)
	defer cancel()
	select {
	case <-c.done:
		return errors.Join(c.errs...)
	case <-ctx.Done():
		c.clear()
		return ctx.Err()
	}
}
