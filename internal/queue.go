package internal

type (
	pointer uint32

	element[T any] struct {
		prev, addr, next pointer
		Value            T
	}

	Queue[T any] struct {
		head, tail pointer      // 头尾指针
		length     int          // 长度
		stack      stack        // 回收站
		elements   []element[T] // 元素列表
		template   element[T]   // 空值模板
	}
)

func NewQueue[T any](capacity uint32) *Queue[T] {
	return &Queue[T]{elements: make([]element[T], 1, 1+capacity)}
}

func (c *Queue[T]) get(addr pointer) *element[T] {
	if addr > 0 {
		return &(c.elements[addr])
	}
	return nil
}

func (c *Queue[T]) getElement() *element[T] {
	if c.stack.Len() > 0 {
		addr := c.stack.Pop()
		v := c.get(addr)
		v.addr = addr
		return v
	}

	addr := pointer(len(c.elements))
	c.elements = append(c.elements, c.template)
	v := c.get(addr)
	v.addr = addr
	return v
}

func (c *Queue[T]) Len() int {
	return c.length
}

func (c *Queue[T]) Front() (value T) {
	if c.head > 0 {
		return c.get(c.head).Value
	}
	return value
}

func (c *Queue[T]) Push(value T) {
	ele := c.getElement()
	ele.Value = value
	c.length++

	if c.tail != 0 {
		tail := c.get(c.tail)
		tail.next = ele.addr
		ele.prev = tail.addr
		c.tail = ele.addr
		return
	}

	c.head = ele.addr
	c.tail = ele.addr
}

func (c *Queue[T]) Pop() (value T) {
	ele := c.get(c.head)
	if ele == nil {
		return value
	}

	c.head = ele.next
	if head := c.get(c.head); head != nil {
		head.prev = 0
	}

	c.length--
	if c.length == 0 {
		c.tail = 0
	}

	c.stack.Push(ele.addr)
	value = ele.Value
	*ele = c.template
	return value
}

func (c *Queue[T]) Range(f func(v T) bool) {
	for i := c.get(c.head); i != nil; i = c.get(i.next) {
		if !f(i.Value) {
			break
		}
	}
}

type stack []pointer

func (c *stack) Len() int {
	return len(*c)
}

func (c *stack) Push(v pointer) {
	*c = append(*c, v)
}

func (c *stack) Pop() pointer {
	var n = c.Len()
	var v = (*c)[n-1]
	*c = (*c)[:n-1]
	return v
}
