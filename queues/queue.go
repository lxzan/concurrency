package queues

import (
	"sync"
	"sync/atomic"
)

const defaultConcurrency = 8

var DefaultQueue = New(WithConcurrency(64))

type (
	options struct {
		concurrency int64
		caller      Caller
	}

	Caller func(f func())

	queue struct {
		mu             *sync.Mutex // 锁
		q              []Job       // 任务队列
		maxConcurrency int64       // 最大并发
		curConcurrency int64       // 当前并发
		caller         Caller      // 异常处理
	}

	Queues struct {
		options *options
		serial  int64
		qs      []*queue
	}
)

func New(opts ...Option) *Queues {
	o := &options{
		concurrency: defaultConcurrency,
		caller:      func(f func()) { f() },
	}
	for _, f := range opts {
		f(o)
	}

	qs := make([]*queue, o.concurrency)
	for i := int64(0); i < o.concurrency; i++ {
		qs[i] = newQueue(o)
	}
	return &Queues{options: o, qs: qs}
}

func (c *Queues) Push(job Job) {
	index := atomic.AddInt64(&c.serial, 1) & (c.options.concurrency - 1)
	c.qs[index].push(job)
}

// newQueue 创建一个任务队列
func newQueue(o *options) *queue {
	return &queue{
		mu:             &sync.Mutex{},
		maxConcurrency: 1,
		curConcurrency: 0,
		caller:         o.caller,
	}
}

// 获取一个任务
func (c *queue) getJob(delta int64) Job {
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

// 递归地执行任务
func (c *queue) do(job Job) {
	c.caller(job.Do)
	if nextJob := c.getJob(-1); nextJob != nil {
		go c.do(nextJob)
	}
}

// push 追加任务, 有资源空闲的话会立即执行
func (c *queue) push(job Job) {
	c.mu.Lock()
	c.q = append(c.q, job)
	c.mu.Unlock()
	if item := c.getJob(0); item != nil {
		go c.do(item)
	}
}
