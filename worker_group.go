package optimizer

import (
	"context"
	"sync"
	"sync/atomic"
)

type WorkerGroup struct {
	sync.Mutex
	ctx         context.Context // context
	collector   *errorCollector // error collector
	q           *Queue          // task queue
	concurrency int64           // max concurrent coroutine
	taskDone    int64           // completed tasks
	taskTotal   int64           // total tasks
	signal      chan bool
	OnMessage   func(options interface{}) error
}

// NewWorkerGroup 新建一个任务集
// concurrency: 最大并发协程数量
func NewWorkerGroup(ctx context.Context, concurrency int64) *WorkerGroup {
	if concurrency <= 0 {
		concurrency = 8
	}
	o := &WorkerGroup{
		ctx:         ctx,
		collector:   &errorCollector{mu: &sync.RWMutex{}},
		q:           NewQueue(),
		concurrency: concurrency,
		taskDone:    0,
		signal:      make(chan bool),
	}
	return o
}

// Len 获取队列中剩余任务数量
func (c *WorkerGroup) Len() int {
	return c.q.Len()
}

// Push 往任务队列中追加任务
func (c *WorkerGroup) Push(eles ...interface{}) {
	atomic.AddInt64(&c.taskTotal, int64(len(eles)))
	c.q.Push(eles...)
}

func (c *WorkerGroup) do() {
	if atomic.LoadInt64(&c.taskDone) == atomic.LoadInt64(&c.taskTotal) {
		c.signal <- true
		return
	}

	if item := c.q.Front(); item != nil {
		go func(doc interface{}) {
			if !isCanceled(c.ctx) {
				if err := c.OnMessage(doc); err != nil {
					c.collector.MarkFailedWithError(err)
				} else {
					c.collector.MarkSucceed()
				}
			}
			atomic.AddInt64(&c.taskDone, 1)
			c.do()
		}(item)
	}
}

// StartAndWait 启动并等待所有任务执行完成
func (c *WorkerGroup) StartAndWait() {
	var taskTotal = atomic.LoadInt64(&c.taskTotal)
	if taskTotal == 0 {
		return
	}

	var co = min(c.concurrency, taskTotal)
	for i := int64(0); i < co; i++ {
		c.do()
	}

	<-c.signal
}

// Err 获取错误返回
func (c *WorkerGroup) Err() error {
	return c.collector.Err()
}
