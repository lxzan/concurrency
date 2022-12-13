package concurrency

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewWorkerQueue(t *testing.T) {
	as := assert.New(t)

	t.Run("sum", func(t *testing.T) {
		var threads = int64(8)
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(1000)
		w := NewWorkerQueue(WithConcurrency(threads))
		for i := 1; i <= 1000; i++ {
			w.AddJob(Job{
				Args: int64(i),
				Do: func(args interface{}) error {
					atomic.AddInt64(&val, args.(int64))
					wg.Done()
					as.LessOrEqual(w.getCurConcurrency(), threads)
					return nil
				},
			})
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		var err error
		var wg = sync.WaitGroup{}
		wg.Add(1)
		w := NewWorkerQueue(WithRecovery())
		w.AddJob(Job{
			Args: nil,
			Do: func(args interface{}) error {
				panic("test")
			},
		})
		w.OnError = func(e error) {
			err = e
			wg.Done()
		}
		wg.Wait()
		as.Error(err)
	})

	t.Run("error", func(t *testing.T) {
		var wg = sync.WaitGroup{}
		wg.Add(1)
		w := NewWorkerQueue()
		w.AddJob(Job{
			Args: nil,
			Do: func(args interface{}) error {
				defer wg.Done()
				return errors.New("internal error")
			},
		})
		wg.Wait()
	})
}
