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
			job := ParameterizedJob(int64(i), func(args interface{}) {
				atomic.AddInt64(&val, args.(int64))
				wg.Done()
			})
			w.Push(job)
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		w := New(WithRecovery())
		job := FuncJob(func() {
			panic("test")
		})
		w.Push(job)
	})

	t.Run("stop", func(t *testing.T) {
		cc := New()
		job := FuncJob(func() {
			time.Sleep(time.Second)
		})
		cc.Push(job)
	})
}
