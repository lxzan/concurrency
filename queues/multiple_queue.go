package queues

import (
	"context"
	"sync/atomic"
	"time"
)

type multipleQueue struct {
	options *options       // 参数
	serial  int64          // 序列号
	size    int64          // 队列大小
	stopped atomic.Uint32  // 是否关闭
	qs      []*singleQueue // 子队列
}

// 创建多重队列
func newMultipleQueue(o *options) *multipleQueue {
	qs := make([]*singleQueue, o.size)
	for i := int64(0); i < o.size; i++ {
		qs[i] = newSingleQueue(o)
	}
	return &multipleQueue{options: o, qs: qs, size: o.size}
}

// Push 追加任务
func (c *multipleQueue) Push(job Job) {
	if c.stopped.Load() == 0 {
		index := atomic.AddInt64(&c.serial, 1) & (c.size - 1)
		c.qs[index].Push(job)
	}
}

// Stop 停止
// 可能需要等待一段时间, 直到所有任务执行完成或者超时
func (c *multipleQueue) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.options.timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if c.doStop(false) {
				return nil
			}
		case <-ctx.Done():
			c.doStop(true)
			return ErrStopTimeout
		}
	}
}

func (c *multipleQueue) doStop(force bool) bool {
	if force {
		c.stopped.Store(1)
		return true
	}

	sum := 0
	for _, item := range c.qs {
		sum += item.len()
	}
	if sum == 0 {
		c.stopped.Store(1)
		return true
	}

	return false
}
