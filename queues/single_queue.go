package queues

import (
	"context"
	"github.com/lxzan/concurrency/logs"
	"sync"
	"time"
)

// 创建一条任务队列
func newSingleQueue(o *options) *singleQueue {
	return &singleQueue{
		mu:             &sync.Mutex{},
		maxConcurrency: o.concurrency,
		curConcurrency: 0,
		caller:         o.caller,
		logger:         o.logger,
		timeout:        o.timeout,
	}
}

type singleQueue struct {
	mu             *sync.Mutex   // 锁
	q              []Job         // 任务队列
	maxConcurrency int64         // 最大并发
	curConcurrency int64         // 当前并发
	caller         Caller        // 异常处理
	logger         logs.Logger   // 日志
	timeout        time.Duration // 退出超时
	stopped        bool          // 是否关闭
}

func (c *singleQueue) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if c.doStop() {
				return
			}
		case <-ctx.Done():
			c.doStop()
			return
		}
	}
}

func (c *singleQueue) doStop() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.q) == 0 {
		c.stopped = true
		return true
	}
	return false
}

// 获取一个任务
func (c *singleQueue) getJob(delta int64) Job {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.curConcurrency += delta
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

// 循环执行任务
func (c *singleQueue) do(job Job) {
	for job != nil {
		c.caller(c.logger, job)
		job = c.getJob(-1)
	}
}

// push 追加任务, 有资源空闲的话会立即执行
func (c *singleQueue) Push(job Job) {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}

	c.q = append(c.q, job)
	c.mu.Unlock()
	if item := c.getJob(0); item != nil {
		go c.do(item)
	}
}
