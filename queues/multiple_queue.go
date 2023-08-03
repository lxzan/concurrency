package queues

import (
	"context"
	"github.com/lxzan/concurrency/internal"
	"github.com/lxzan/concurrency/logs"
	"sync"
	"sync/atomic"
	"time"
)

type (
	multipleQueue struct {
		options *options
		serial  int64
		size    int64
		qs      []*multipleQueueChild
	}

	multipleQueueChild struct {
		mu             *sync.Mutex // 锁
		q              []Job       // 任务队列
		maxConcurrency int64       // 最大并发
		curConcurrency int64       // 当前并发
		caller         Caller      // 异常处理
		logger         logs.Logger // 日志
	}
)

// 创建N条并发度为1的任务队列
func newMultipleQueue(o *options) *multipleQueue {
	size := internal.ToBinaryNumber(o.concurrency)
	qs := make([]*multipleQueueChild, size)
	for i := int64(0); i < size; i++ {
		qs[i] = &multipleQueueChild{
			mu:             &sync.Mutex{},
			maxConcurrency: 1,
			curConcurrency: 0,
			caller:         o.caller,
			logger:         o.logger,
		}
	}
	return &multipleQueue{options: o, qs: qs, size: size}
}

// Push 追加任务
func (c *multipleQueue) Push(job Job) {
	index := atomic.AddInt64(&c.serial, 1) & (c.size - 1)
	c.qs[index].push(job)
}

// Stop 停止
// 可能需要等待一段时间, 直到所有任务执行完成或者超时
func (c *multipleQueue) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), c.options.timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			sum := 0
			for _, item := range c.qs {
				sum += item.len()
			}
			if sum == 0 {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *multipleQueueChild) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.q)
}

// 获取一个任务
func (c *multipleQueueChild) getJob(delta int64) Job {
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
func (c *multipleQueueChild) do(job Job) {
	for job != nil {
		c.caller(c.logger, job)
		job = c.getJob(-1)
	}
}

// push 追加任务, 有资源空闲的话会立即执行
func (c *multipleQueueChild) push(job Job) {
	c.mu.Lock()
	c.q = append(c.q, job)
	c.mu.Unlock()
	if item := c.getJob(0); item != nil {
		go c.do(item)
	}
}
