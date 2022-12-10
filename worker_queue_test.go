package concurrency

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewWorkerQueue(t *testing.T) {
	as := assert.New(t)
	t.Run("sum", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(100)
		w := NewWorkerQueue(WithConcurrency(8))
		for i := 1; i <= 100; i++ {
			w.Push(Job{
				Args: int64(i),
				Do: func(args interface{}) error {
					println(atomic.LoadInt64(&w.curConcurrency))
					atomic.AddInt64(&val, args.(int64))
					wg.Done()
					return nil
				},
			})
		}
		wg.Wait()
		as.Equal(int64(5050), val)
	})

	t.Run("recover", func(t *testing.T) {
		var wg = sync.WaitGroup{}
		wg.Add(1)
		w := NewWorkerQueue(WithRecovery())
		w.Push(Job{
			Args: nil,
			Do: func(args interface{}) error {
				wg.Done()
				panic("test")
			},
		})
		w.OnError = func(err error) {
			as.Error(err)
		}
		wg.Wait()
	})
}
