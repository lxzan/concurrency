package concurrency

import "sync"

type Queue struct {
	sync.Mutex
	data []interface{}
}

func NewQueue() *Queue {
	return &Queue{
		data: make([]interface{}, 0),
	}
}

func (c *Queue) Push(eles ...interface{}) {
	c.Lock()
	c.data = append(c.data, eles...)
	c.Unlock()
}

func (c *Queue) Front() interface{} {
	c.Lock()
	defer c.Unlock()

	if n := len(c.data); n == 0 {
		return nil
	} else {
		var result = c.data[0]
		c.data = c.data[1:]
		return result
	}
}

func (c *Queue) Len() int {
	c.Lock()
	length := len(c.data)
	c.Unlock()
	return length
}

// All 返回所有数据并清空队列
func (c *Queue) All() []interface{} {
	c.Lock()
	data := c.data
	c.data = make([]interface{}, 0)
	c.Unlock()
	return data
}
