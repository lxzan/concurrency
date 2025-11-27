package queues

import (
	"context"
	"sync"
	"sync/atomic"
)

type (
	multipleQueue struct {
		conf    *options       // 参数
		serial  atomic.Int64   // 序列号
		stopped bool           // 是否关闭
		qs      []*singleQueue // 子队列
	}

	errWrapper struct{ err error }
)

// 创建多重队列
func newMultipleQueue(o *options) *multipleQueue {
	qs := make([]*singleQueue, o.sharding)
	for i := int64(0); i < o.sharding; i++ {
		qs[i] = newSingleQueue(o)
	}
	return &multipleQueue{conf: o, qs: qs}
}

func (c *multipleQueue) Len() int {
	var sum = 0
	for _, q := range c.qs {
		sum += q.Len()
	}
	return sum
}

// Push 追加任务
func (c *multipleQueue) Push(job Job) {
	i := c.serial.Add(1) & (c.conf.sharding - 1)
	c.qs[i].Push(job)
}

// Stop 停止
// 可能需要等待一段时间, 直到所有任务执行完成或者超时
func (c *multipleQueue) Stop(ctx context.Context) error {
	var err = atomic.Pointer[errWrapper]{}
	var wg = sync.WaitGroup{}
	wg.Add(int(c.conf.sharding))
	for i, _ := range c.qs {
		go func(q *singleQueue) {
			err.CompareAndSwap(nil, &errWrapper{q.Stop(ctx)})
			wg.Done()
		}(c.qs[i])
	}
	wg.Wait()
	if e := err.Load(); e != nil {
		return e.err
	}
	return nil
}
