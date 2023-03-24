package concurrency

import (
	"context"
	"errors"
	"github.com/hashicorp/go-multierror"
	"sync"
	"time"
)

const (
	defaultConcurrency = 8                // 默认并发度
	defaultWaitTimeout = 60 * time.Second // 线程同步等待超时
)

var ErrWaitTimeout = errors.New("wait timeout")

type WorkerGroup[T any] struct {
	mu          *sync.Mutex        // 锁
	concurrency int64              // 并发
	timeout     time.Duration      // 超时
	err         *multierror.Error  // 错误
	done        chan bool          // 信号
	q           []T                // 任务队列
	taskDone    int64              // 已完成任务数量
	taskTotal   int64              // 总任务数量
	OnMessage   func(args T) error // 任务处理
}

// NewWorkerGroup 新建一个任务集
func NewWorkerGroup[T any]() *WorkerGroup[T] {
	c := &WorkerGroup[T]{
		err:         &multierror.Error{},
		concurrency: defaultConcurrency,
		timeout:     defaultWaitTimeout,
		mu:          &sync.Mutex{},
		q:           make([]T, 0),
		taskDone:    0,
		done:        make(chan bool),
	}
	c.OnMessage = func(args T) error {
		return nil
	}
	return c
}

func (c *WorkerGroup[T]) SetConcurrency(num int64) *WorkerGroup[T] {
	if num >= 0 {
		c.concurrency = num
	}
	return c
}

func (c *WorkerGroup[T]) SetTimeout(d time.Duration) *WorkerGroup[T] {
	if d >= 0 {
		c.timeout = d
	}
	return c
}

func (c *WorkerGroup[T]) clear() {
	c.mu.Lock()
	c.q = c.q[:0]
	c.mu.Unlock()
}

func (c *WorkerGroup[T]) getJob() (v T, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n := len(c.q); n == 0 {
		return
	}
	var result = c.q[0]
	c.q = c.q[1:]
	return result, true
}

func (c *WorkerGroup[T]) appendError(err error) {
	c.mu.Lock()
	c.err = multierror.Append(c.err, err)
	c.mu.Unlock()
}

// incrAndIsDone
// 已完成任务+1, 并检查任务是否全部完成
func (c *WorkerGroup[T]) incrAndIsDone() bool {
	c.mu.Lock()
	c.taskDone++
	ok := c.taskDone == c.taskTotal
	c.mu.Unlock()
	return ok
}

func (c *WorkerGroup[T]) do(args T) {
	if err := recoveryCaller(args, c.OnMessage); err != nil {
		c.appendError(err)
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
func (c *WorkerGroup[T]) Len() int {
	c.mu.Lock()
	x := len(c.q)
	c.mu.Unlock()
	return x
}

// AddJob 往任务队列中追加任务
func (c *WorkerGroup[T]) Push(eles ...T) {
	c.mu.Lock()
	c.taskTotal += int64(len(eles))
	c.q = append(c.q, eles...)
	c.mu.Unlock()
}

// Update 线程安全操作
func (c *WorkerGroup[T]) Update(f func()) {
	c.mu.Lock()
	f()
	c.mu.Unlock()
}

// Start 启动并等待所有任务执行完成
func (c *WorkerGroup[T]) Start() error {
	var taskTotal = int64(c.Len())
	if taskTotal == 0 {
		return nil
	}

	var co = min(c.concurrency, taskTotal)
	for i := int64(0); i < co; i++ {
		if item, ok := c.getJob(); ok {
			go c.do(item)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	select {
	case <-c.done:
		return c.err.ErrorOrNil()
	case <-ctx.Done():
		c.clear()
		return ErrWaitTimeout
	}
}
