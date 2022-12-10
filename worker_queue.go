package concurrency

import (
	"context"
	"sync"
	"time"
)

type WorkerQueue struct {
	mu             *sync.Mutex
	config         *Config
	q              []Job
	maxConcurrency int64
	curConcurrency int64
	OnError        func(err error)
}

// NewWorkerQueue 创建一个工作队列
func NewWorkerQueue(options ...Option) *WorkerQueue {
	config := &Config{}
	for _, fn := range options {
		fn(config)
	}
	return &WorkerQueue{
		mu:             &sync.Mutex{},
		config:         config.init(),
		q:              make([]Job, 0),
		maxConcurrency: config.Concurrency,
		curConcurrency: 0,
	}
}

// Push 追加任务, 有资源空闲的话会立即执行
func (c *WorkerQueue) Push(jobs ...Job) {
	c.mu.Lock()
	c.q = append(c.q, jobs...)
	c.mu.Unlock()

	var n = len(jobs)
	for i := 0; i < n; i++ {
		c.do()
	}
}

func (c *WorkerQueue) getJob() interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.curConcurrency >= c.maxConcurrency {
		return nil
	}
	if n := len(c.q); n == 0 {
		return nil
	}

	var result = c.q[0]
	c.q = c.q[1:]
	c.curConcurrency++
	return result
}

func (c *WorkerQueue) incr(d int64) {
	c.mu.Lock()
	c.curConcurrency += d
	c.mu.Unlock()
}

func (c *WorkerQueue) do() {
	if item := c.getJob(); item != nil {
		go func(job Job) {
			if !isCanceled(c.config.Context) {
				c.callOnError(c.config.Caller(job))
			}
			c.incr(-1)
			c.do()
		}(item.(Job))
	}
}

func (c *WorkerQueue) callOnError(err error) {
	if err == nil {
		return
	}
	if c.OnError != nil {
		c.OnError(err)
	}
}

func (c *WorkerQueue) getCurConcurrency() int64 {
	c.mu.Lock()
	x := c.curConcurrency
	c.mu.Unlock()
	return x
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
			c.mu.Lock()
			x := int64(len(c.q)) + c.curConcurrency
			c.mu.Unlock()
			if x == 0 {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
