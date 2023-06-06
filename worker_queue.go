package concurrency

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

var DefaultQueue = NewWorkerQueue().SetConcurrency(64)

type WorkerQueue struct {
	mu             *sync.Mutex                 // 锁
	q              []Job                       // 任务队列
	maxConcurrency int64                       // 最大并发
	curConcurrency int64                       // 当前并发
	OnRecovery     func(exception interface{}) // 异常处理
}

// NewWorkerQueue 创建一个任务队列
func NewWorkerQueue() *WorkerQueue {
	return &WorkerQueue{
		mu:             &sync.Mutex{},
		maxConcurrency: defaultConcurrency,
		curConcurrency: 0,
		OnRecovery:     func(exception interface{}) {},
	}
}

func (c *WorkerQueue) SetConcurrency(num int64) *WorkerQueue {
	if num >= 0 {
		c.maxConcurrency = num
	}
	return c
}

// 获取一个任务
func (c *WorkerQueue) getJob(delta int64) Job {
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
func (c *WorkerQueue) do(job Job) {
	for job != nil {
		c.call(job)
		job = c.getJob(-1)
	}
}

// Push 追加任务, 有资源空闲的话会立即执行
func (c *WorkerQueue) Push(job Job) {
	c.mu.Lock()
	c.q = append(c.q, job)
	c.mu.Unlock()
	if item := c.getJob(0); item != nil {
		go c.do(item)
	}
}

func (c *WorkerQueue) call(job Job) {
	defer func() {
		if fatalError := recover(); fatalError != nil {
			var msg = make([]byte, 0, 256)
			msg = append(msg, fmt.Sprintf("fatal error: %v\n", fatalError)...)
			for i := 1; true; i++ {
				_, caller, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				if !strings.Contains(caller, "src/runtime") {
					msg = append(msg, fmt.Sprintf("caller: %s, line: %d\n", caller, line)...)
				}
			}
			c.OnRecovery(msg)
		}
	}()
	job.Do()
}

// Len 获取队列中剩余任务数量
func (c *WorkerQueue) Len() int {
	c.mu.Lock()
	x := len(c.q)
	c.mu.Unlock()
	return x
}

func (c *WorkerQueue) Stop(timeout time.Duration) {
	ticker := time.NewTicker(10 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			num := int(c.curConcurrency) + len(c.q)
			c.mu.Unlock()
			if num == 0 {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
