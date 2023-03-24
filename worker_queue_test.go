package concurrency

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
		w := NewWorkerQueue().SetConcurrency(16)
		for i := 1; i <= 1000; i++ {
			job := ParameterizedJob(int64(i), func(args interface{}) {
				atomic.AddInt64(&val, args.(int64))
				wg.Done()
				as.LessOrEqual(w.curConcurrency, w.maxConcurrency)
			})
			w.Push(job)
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		w := NewWorkerQueue()
		job := FuncJob(func() {
			panic("test")
		})
		w.Push(job)
		w.Stop(time.Second)
	})

	t.Run("stop", func(t *testing.T) {
		cc := NewWorkerQueue()
		job := FuncJob(func() {
			println(cc.Len())
			time.Sleep(time.Second)
		})
		cc.Push(job)
		cc.Stop(100 * time.Millisecond)
	})
}
