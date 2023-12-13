package queues

import (
	"context"
	"github.com/lxzan/concurrency/internal"
	"sync"
	"time"
)

// 创建一条任务队列
func newSingleQueue(o *options) *singleQueue {
	return &singleQueue{
		mu:             &sync.Mutex{},
		conf:           o,
		maxConcurrency: int32(o.concurrency),
		curConcurrency: 0,
		q:              internal.NewQueue[Job](8),
	}
}

type singleQueue struct {
	mu             *sync.Mutex // 锁
	conf           *options
	q              *internal.Queue[Job] // 任务队列
	maxConcurrency int32                // 最大并发
	curConcurrency int32                // 当前并发
	stopped        bool                 // 是否关闭
}

func (c *singleQueue) Stop(ctx context.Context) error {
	if !c.cas(false, true) {
		return nil
	}

	ctx1, cancel := context.WithTimeout(ctx, c.conf.timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if c.finish() {
				return nil
			}
		case <-ctx1.Done():
			if c.finish() {
				return nil
			}
			return ctx1.Err()
		}
	}
}

// 获取一个任务
func (c *singleQueue) getJob(newJob Job, delta int32) Job {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.stopped && newJob != nil {
		c.q.Push(newJob)
	}
	c.curConcurrency += delta
	if c.curConcurrency >= c.maxConcurrency {
		return nil
	}
	if job := c.q.Pop(); job != nil {
		c.curConcurrency++
		return job
	}
	return nil
}

// 循环执行任务
func (c *singleQueue) do(job Job) {
	for job != nil {
		c.conf.caller(c.conf.logger, job)
		job = c.getJob(nil, -1)
	}
}

// Push 追加任务, 有资源空闲的话会立即执行
func (c *singleQueue) Push(job Job) {
	if nextJob := c.getJob(job, 0); nextJob != nil {
		go c.do(nextJob)
	}
}

func (c *singleQueue) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.q.Len()
}

func (c *singleQueue) finish() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.q.Len()+int(c.curConcurrency) == 0
}

func (c *singleQueue) cas(old, new bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped == old {
		c.stopped = new
		return true
	}
	return false
}
