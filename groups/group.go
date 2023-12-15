package groups

import (
	"context"
	"errors"
	"github.com/lxzan/concurrency/internal"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultConcurrency = 8                // 默认并发度
	defaultWaitTimeout = 60 * time.Second // 默认线程同步等待超时
)

var defaultCaller Caller = func(args any, f func(any) error) error { return f(args) }

type (
	Caller func(args any, f func(any) error) error

	Group[T any] struct {
		options    *options           // 配置
		mu         sync.Mutex         // 锁
		ctx        context.Context    // 上下文
		cancelFunc context.CancelFunc // 取消函数
		canceled   atomic.Uint32      // 是否已取消
		errs       []error            // 错误
		done       chan bool          // 完成信号
		q          []T                // 任务队列
		taskDone   int64              // 已完成任务数量
		taskTotal  int64              // 总任务数量
		OnMessage  func(args T) error // 任务处理
		OnError    func(err error)    // 错误处理
	}
)

// New 新建一个任务集
func New[T any](opts ...Option) *Group[T] {
	o := new(options)
	opts = append(opts, withInitialize())
	for _, f := range opts {
		f(o)
	}

	c := &Group[T]{
		options:  o,
		q:        make([]T, 0),
		taskDone: 0,
		done:     make(chan bool),
	}
	c.ctx, c.cancelFunc = context.WithTimeout(context.Background(), o.timeout)
	c.OnMessage = func(args T) error {
		return nil
	}
	c.OnError = func(err error) {}

	return c
}

func (c *Group[T]) clearJob() {
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

func (c *Group[T]) getError() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return errors.Join(c.errs...)
}

func (c *Group[T]) jobFunc(v any) error {
	if c.canceled.Load() == 1 {
		return nil
	}
	return c.OnMessage(v.(T))
}

func (c *Group[T]) do(args T) {
	if err := c.options.caller(args, c.jobFunc); err != nil {
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

// Cancel 取消队列中剩余任务的执行
func (c *Group[T]) Cancel() {
	if c.canceled.CompareAndSwap(0, 1) {
		c.cancelFunc()
	}
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

	defer c.cancelFunc()

	select {
	case <-c.done:
		return c.getError()
	case <-c.ctx.Done():
		c.clearJob()
		return c.ctx.Err()
	}
}
