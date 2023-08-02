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

	t.Run("stop", func(t *testing.T) {
		cc := New()
		cc.Push(func() {
			time.Sleep(time.Second)
		})
	})
}
