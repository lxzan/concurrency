package concurrency

import (
	"context"
	"sync/atomic"
	"time"
)

var DefaultWorker = NewWorkerQueue(context.Background(), 16)

type (
	WorkerQueue struct {
		ctx            context.Context
		q              *Queue
		maxConcurrency int64
		curConcurrency int64
		OnError        func(err error)
	}

	Job struct {
		Args interface{}
		Do   func(args interface{}) error
	}
)

// NewWorkerQueue 创建一个工作队列
// concurrency 最大并发协程数量
func NewWorkerQueue(ctx context.Context, threads int64) *WorkerQueue {
	return &WorkerQueue{
		ctx:            ctx,
		q:              NewQueue(),
		maxConcurrency: threads,
		curConcurrency: 0,
	}
}

// Push 追加任务, 有资源空闲的话会立即执行
func (c *WorkerQueue) Push(jobs ...Job) {
	for i, _ := range jobs {
		c.q.Push(jobs[i])
		c.do()
	}
}

func (c *WorkerQueue) do() {
	if atomic.LoadInt64(&c.curConcurrency) >= c.maxConcurrency {
		return
	}

	item := c.q.Front()
	if item == nil {
		return
	}

	atomic.AddInt64(&c.curConcurrency, 1)
	go func(job Job) {
		if !isCanceled(c.ctx) {
			c.callOnError(job.Do(job.Args))
		}
		atomic.AddInt64(&c.curConcurrency, -1)
		c.do()
	}(item.(Job))
}

func (c *WorkerQueue) callOnError(err error) {
	if err == nil {
		return
	}
	if c.OnError != nil {
		c.OnError(err)
	}
}

// Stop 优雅退出
// timeout 超时时间
func (c *WorkerQueue) StopAndWait(timeout time.Duration) {
	ticker := time.NewTicker(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ticker.C:
			if c.q.Len() == 0 && atomic.LoadInt64(&c.curConcurrency) == 0 {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
