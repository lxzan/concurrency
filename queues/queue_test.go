package queues

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerQueue(t *testing.T) {
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
		w := New(WithRecovery())
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
}
