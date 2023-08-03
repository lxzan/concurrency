package queues

import (
	"github.com/lxzan/concurrency/logs"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingleQueue(t *testing.T) {
	as := assert.New(t)

	t.Run("sum", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(1000)
		w := New(WithConcurrency(16))
		for i := 1; i <= 1000; i++ {
			args := int64(i)
			w.Push(func() {
				atomic.AddInt64(&val, args)
				wg.Done()
			})
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		w := New(WithRecovery(), WithLogger(logs.DefaultLogger))
		w.Push(func() {
			panic("test")
		})
	})

	t.Run("stop timeout", func(t *testing.T) {
		cc := New(WithTimeout(time.Millisecond))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Second)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop()
		assert.Equal(t, int64(0), atomic.LoadInt64(&sum))
	})

	t.Run("stop graceful", func(t *testing.T) {
		cc := New(WithTimeout(time.Second))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop()
		assert.Equal(t, int64(1), atomic.LoadInt64(&sum))
	})

	t.Run("", func(t *testing.T) {
		q := New(WithConcurrency(1))
		q.Push(func() { time.Sleep(100 * time.Millisecond) })
		q.Push(func() {})
		q.Stop()
		q.Push(func() {})
		assert.Equal(t, 0, q.(*singleQueue).len())
	})
}

func TestMultiQueue(t *testing.T) {
	as := assert.New(t)

	t.Run("sum", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(1000)
		w := New(WithConcurrency(16), WithMultiple(8))
		for i := 1; i <= 1000; i++ {
			args := int64(i)
			w.Push(func() {
				atomic.AddInt64(&val, args)
				wg.Done()
			})
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		w := New(WithRecovery(), WithLogger(logs.DefaultLogger), WithMultiple(8))
		w.Push(func() {
			panic("test")
		})
	})

	t.Run("stop timeout", func(t *testing.T) {
		cc := New(WithTimeout(time.Millisecond), WithMultiple(8))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Second)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop()
		assert.Equal(t, int64(0), atomic.LoadInt64(&sum))
	})

	t.Run("stop graceful", func(t *testing.T) {
		cc := New(WithTimeout(time.Second), WithMultiple(8))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop()
		assert.Equal(t, int64(1), atomic.LoadInt64(&sum))
	})

	t.Run("", func(t *testing.T) {
		q := New(WithConcurrency(1), WithMultiple(1))
		q.Push(func() { time.Sleep(100 * time.Millisecond) })
		q.Push(func() {})
		q.Stop()
		q.Push(func() {})
	})
}
