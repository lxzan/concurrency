package internal

import (
	"container/list"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestQueue_Range(t *testing.T) {
	const count = 1000
	var q = NewQueue[int](0)
	var a []int
	for i := 0; i < count; i++ {
		v := rand.Intn(count)
		q.Push(v)
		a = append(a, v)
	}

	assert.Equal(t, q.Len(), count)

	var b []int
	q.Range(func(v int) bool {
		b = append(b, v)
		return len(b) < 100
	})
	assert.Equal(t, len(b), 100)

	var i = 0
	for q.Len() > 0 {
		v := q.Pop()
		assert.Equal(t, a[i], v)
		i++
	}
}

func TestQueue_Addr(t *testing.T) {
	const count = 1000
	var q = NewQueue[int](0)
	for i := 0; i < count; i++ {
		v := rand.Intn(count)
		if v&7 == 0 {
			q.Pop()
		} else {
			q.Push(v)
		}
	}

	var sum = 0
	for i := q.get(q.head); i != nil; i = q.get(i.next) {
		sum++

		prev := q.get(i.prev)
		next := q.get(i.next)
		if prev != nil {
			assert.Equal(t, prev.next, i.addr)
		}
		if next != nil {
			assert.Equal(t, i.addr, next.prev)
		}
	}

	assert.Equal(t, q.Len(), sum)
	if head := q.get(q.head); head != nil {
		assert.Zero(t, head.prev)
	}
	if tail := q.get(q.tail); tail != nil {
		assert.Zero(t, tail.next)
	}
}

func TestQueue_Pop(t *testing.T) {
	var q = NewQueue[int](0)
	assert.Zero(t, q.Front())
	assert.Zero(t, q.Pop())

	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Pop()
	q.Push(4)
	q.Push(5)
	q.Pop()

	var arr []int
	q.Range(func(v int) bool {
		arr = append(arr, v)
		return true
	})
	assert.Equal(t, q.Front(), 3)
	assert.True(t, IsSameSlice(arr, []int{3, 4, 5}))
	assert.Equal(t, len(q.elements), 5)
	assert.Equal(t, q.stack.Len(), 1)
}

func TestQueue_Random(t *testing.T) {
	var count = 10000
	var q = NewQueue[int](0)
	var linkedlist = list.New()
	for i := 0; i < count; i++ {
		var flag = rand.Intn(4)
		var val = rand.Int()
		switch flag {
		case 0, 1, 2:
			q.Push(val)
			linkedlist.PushBack(val)
		default:
			if q.Len() > 0 {
				q.Pop()
				linkedlist.Remove(linkedlist.Front())
			}
		}
	}

	for i := linkedlist.Front(); i != nil; i = i.Next() {
		var val = q.Pop()
		assert.Equal(t, i.Value, val)
	}
}

func BenchmarkQueue_PushAndPop(b *testing.B) {
	const count = 1000
	var q = NewQueue[int](count)
	for i := 0; i < b.N; i++ {
		for j := 0; j < count/4; j++ {
			q.Push(j)
		}
		for j := 0; j < count/4; j++ {
			q.Pop()
		}
		for j := 0; j < count/4; j++ {
			q.Push(j)
		}
		for j := 0; j < count/4; j++ {
			q.Pop()
		}
	}
}
