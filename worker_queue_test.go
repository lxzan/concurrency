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
	t.Run("", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(4)
		w := NewWorkerQueue()
		job1 := Job{
			Args: 1,
			Do: func(args interface{}) error {
				atomic.AddInt64(&val, 1)
				wg.Done()
				return nil
			},
		}
		job2 := Job{
			Args: struct{}{},
			Do: func(args interface{}) error {
				atomic.AddInt64(&val, 2)
				wg.Done()
				return nil
			},
		}
		job3 := Job{
			Args: "",
			Do: func(args interface{}) error {
				atomic.AddInt64(&val, 4)
				wg.Done()
				return nil
			},
		}
		w.Push(job1, job2, job3, job2)
		wg.Wait()
		as.Equal(int64(9), val)
	})

	t.Run("", func(t *testing.T) {
		val := int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(4)
		w := NewWorkerQueue()
		job1 := Job{
			Args: 1,
			Do: func(args interface{}) error {
				atomic.AddInt64(&val, 1)
				wg.Done()
				return nil
			},
		}
		job2 := Job{
			Args: struct{}{},
			Do: func(args interface{}) error {
				atomic.AddInt64(&val, 4)
				wg.Done()
				return nil
			},
		}

		w.Push(job1, job1)
		time.Sleep(100 * time.Millisecond)
		w.Push(job2, job2)
		wg.Wait()
		as.Equal(int64(10), atomic.LoadInt64(&val))
	})
}
