package queues

import (
	"context"
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
		cc := New(
			WithConcurrency(1),
			WithTimeout(100*time.Millisecond),
		)

		cc.Push(func() {
			time.Sleep(500 * time.Millisecond)
		})
		cc.Push(func() {
			time.Sleep(500 * time.Millisecond)
		})

		err := cc.Stop(context.Background())
		as.Error(err)
	})

	t.Run("stop graceful", func(t *testing.T) {
		cc := New(WithTimeout(time.Second))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop(context.Background())
		assert.Equal(t, int64(1), atomic.LoadInt64(&sum))
	})

	t.Run("", func(t *testing.T) {
		q := New(WithConcurrency(1))
		q.Push(func() { time.Sleep(100 * time.Millisecond) })
		q.Push(func() {})
		q.Stop(context.Background())
		q.Push(func() {})
		assert.Equal(t, 0, q.(*singleQueue).Len())
	})

	t.Run("stop", func(t *testing.T) {
		var q = New()
		var ctx = context.Background()
		q.Stop(ctx)
		q.Stop(ctx)
	})
}

func TestMultiQueue(t *testing.T) {
	as := assert.New(t)

	t.Run("sum", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(1000)
		w := New(WithConcurrency(16), WithSharding(8))
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
		w := New(WithRecovery(), WithLogger(logs.DefaultLogger), WithSharding(8))
		w.Push(func() {
			panic("test")
		})
	})

	t.Run("stop", func(t *testing.T) {
		cc := New(
			WithSharding(2),
			WithConcurrency(1),
		)
		as.Nil(cc.Stop(context.Background()))
	})

	t.Run("stop finished", func(t *testing.T) {
		cc := New(WithSharding(8))

		job := func() {
			time.Sleep(120 * time.Millisecond)
		}
		cc.Push(job)
		cc.Push(job)
		cc.Push(job)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(150 * time.Millisecond)
			cancel()
		}()
		t0 := time.Now()
		err := cc.Stop(ctx)
		println(time.Since(t0).String())
		as.Nil(err)
	})

	t.Run("stop timeout", func(t *testing.T) {
		cc := New(
			WithTimeout(100*time.Millisecond),
			WithSharding(2),
			WithConcurrency(1),
		)
		job := func() {
			time.Sleep(500 * time.Millisecond)
		}
		cc.Push(job)
		cc.Push(job)
		cc.Push(job)
		err := cc.Stop(context.Background())
		as.Error(err)
		as.Equal(cc.Len(), 1)
	})

	t.Run("stop graceful", func(t *testing.T) {
		cc := New(WithTimeout(time.Second), WithSharding(8))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop(context.Background())
		assert.Equal(t, int64(1), atomic.LoadInt64(&sum))
	})

	t.Run("", func(t *testing.T) {
		q := New(WithConcurrency(1))
		q.Push(func() { time.Sleep(100 * time.Millisecond) })
		q.Push(func() {})
		q.Stop(context.Background())
		q.Push(func() {})
	})
}
