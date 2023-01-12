package concurrency

import (
	"github.com/hashicorp/go-multierror"
	"sync"
)

type WorkerGroup struct {
	mu        *sync.Mutex     // 锁
	err       error           // 错误
	config    *Config         // 配置
	done      chan bool       // 信号
	q         []Job           // 任务队列
	taskDone  int64           // 已完成任务数量
	taskTotal int64           // 总任务数量
	OnError   func(err error) // 错误处理函数. 一般用来打印错误; 放弃剩余任务
}

// NewWorkerGroup 新建一个任务集
func NewWorkerGroup(options ...Option) *WorkerGroup {
	config := &Config{}
	for _, fn := range options {
		fn(config)
	}
	o := &WorkerGroup{
		mu:       &sync.Mutex{},
		config:   config.init(),
		q:        make([]Job, 0),
		taskDone: 0,
		done:     make(chan bool),
	}
	return o
}

func (c *WorkerGroup) getJob() interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n := len(c.q); n == 0 {
		return nil
	}
	var result = c.q[0]
	c.q = c.q[1:]
	return result
}

func (c *WorkerGroup) callOnError(err error) {
	if err == nil {
		return
	}
	if c.OnError != nil {
		c.OnError(err)
	}
	c.mu.Lock()
	c.err = multierror.Append(err)
	c.mu.Unlock()
}

// incrAndIsDone
// 已完成任务+1, 并检查任务是否全部完成
func (c *WorkerGroup) incrAndIsDone() bool {
	c.mu.Lock()
	c.taskDone++
	ok := c.taskDone == c.taskTotal
	c.mu.Unlock()
	return ok
}

func (c *WorkerGroup) do(job Job) {
	if !isCanceled(c.config.Context) {
		c.callOnError(c.config.Caller(job))
	}
	if c.incrAndIsDone() {
		c.done <- true
		return
	}
	if nextJob := c.getJob(); nextJob != nil {
		c.do(nextJob.(Job))
	}
}

// Len 获取队列中剩余任务数量
func (c *WorkerGroup) Len() int {
	c.mu.Lock()
	x := len(c.q)
	c.mu.Unlock()
	return x
}

// AddJob 往任务队列中追加任务
func (c *WorkerGroup) AddJob(jobs ...Job) {
	c.mu.Lock()
	c.taskTotal += int64(len(jobs))
	c.q = append(c.q, jobs...)
	c.mu.Unlock()
}

// StartAndWait 启动并等待所有任务执行完成
func (c *WorkerGroup) StartAndWait() error {
	var taskTotal = int64(c.Len())
	if taskTotal == 0 {
		return nil
	}

	var co = min(c.config.Concurrency, taskTotal)
	for i := int64(0); i < co; i++ {
		if item := c.getJob(); item != nil {
			c.callOnError(gpool.Submit(func() { c.do(item.(Job)) }))
		}
	}

	<-c.done
	return c.err
}
