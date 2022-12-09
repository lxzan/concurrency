package concurrency

import (
	"github.com/hashicorp/go-multierror"
	"sync"
	"sync/atomic"
)

type WorkerGroup struct {
	mu        *sync.Mutex
	err       error
	config    *Config
	done      chan bool
	q         *Queue
	taskDone  int64
	taskTotal int64
}

// NewWorkerGroup 新建一个任务集
func NewWorkerGroup(options ...Option) *WorkerGroup {
	config := new(Config).init()
	for _, fn := range options {
		fn(config)
	}
	o := &WorkerGroup{
		mu:       &sync.Mutex{},
		config:   config,
		q:        NewQueue(),
		taskDone: 0,
		done:     make(chan bool),
	}
	return o
}

// Len 获取队列中剩余任务数量
func (c *WorkerGroup) Len() int {
	return c.q.Len()
}

// Push 往任务队列中追加任务
func (c *WorkerGroup) Push(jobs ...Job) {
	atomic.AddInt64(&c.taskTotal, int64(len(jobs)))
	for i, _ := range jobs {
		c.q.Push(jobs[i])
	}
}

func (c *WorkerGroup) appendError(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	c.err = multierror.Append(err)
	c.mu.Unlock()
}

func (c *WorkerGroup) do() {
	if atomic.LoadInt64(&c.taskDone) == atomic.LoadInt64(&c.taskTotal) {
		c.done <- true
		return
	}

	if item := c.q.Front(); item != nil {
		go func(job Job) {
			if !isCanceled(c.config.Context) {
				c.appendError(c.config.Caller(job))
			}
			atomic.AddInt64(&c.taskDone, 1)
			c.do()
		}(item.(Job))
	}
}

// StartAndWait 启动并等待所有任务执行完成
func (c *WorkerGroup) StartAndWait() {
	var taskTotal = atomic.LoadInt64(&c.taskTotal)
	if taskTotal == 0 {
		return
	}

	var co = min(c.config.Concurrency, taskTotal)
	for i := int64(0); i < co; i++ {
		c.do()
	}

	<-c.done
	c.appendError(c.config.err)
}

// Err 获取错误返回
func (c *WorkerGroup) Err() error {
	return c.err
}
