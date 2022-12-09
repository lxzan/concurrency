package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
)

type WorkerGroup struct {
	sync.Mutex
	ctx            context.Context // context
	collector      *errorCollector // error collector
	q              *Queue          // task queue
	taskDone       int64           // completed tasks
	taskTotal      int64           // total tasks
	maxConcurrency int64           // max concurrent coroutine
	curConcurrency int64           // current concurrent coroutine
	done           chan bool
}

// NewWorkerGroup 新建一个任务集
// threads: 最大并发协程数量
func NewWorkerGroup(ctx context.Context, threads int64) *WorkerGroup {
	if threads <= 0 {
		threads = 8
	}
	o := &WorkerGroup{
		ctx:            ctx,
		collector:      &errorCollector{mu: &sync.RWMutex{}},
		q:              NewQueue(),
		maxConcurrency: threads,
		taskDone:       0,
		done:           make(chan bool),
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

func (c *WorkerGroup) do() {
	if atomic.LoadInt64(&c.taskDone) == atomic.LoadInt64(&c.taskTotal) {
		c.done <- true
		return
	}

	if item := c.q.Front(); item != nil {
		go func(job Job) {
			if !isCanceled(c.ctx) {
				if err := job.Do(job.Args); err != nil {
					c.collector.MarkFailedWithError(err)
				} else {
					c.collector.MarkSucceed()
				}
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

	var co = min(c.maxConcurrency, taskTotal)
	for i := int64(0); i < co; i++ {
		c.do()
	}

	<-c.done
}

// Err 获取错误返回
func (c *WorkerGroup) Err() error {
	return c.collector.Err()
}
