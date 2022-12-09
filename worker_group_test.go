package optimizer

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTaskGroup(t *testing.T) {
	as := assert.New(t)
	t.Run("success", func(t *testing.T) {
		mu := sync.Mutex{}
		listA := make([]uint8, 0)
		listB := make([]uint8, 0)
		ctl := NewWorkerGroup(context.Background(), 8)
		for i := 0; i < 100; i++ {
			ctl.Push(uint8(i))
			listB = append(listB, uint8(i))
		}
		ctl.OnMessage = func(options interface{}) error {
			mu.Lock()
			listA = append(listA, options.(uint8))
			mu.Unlock()
			return nil
		}
		ctl.StartAndWait()
		as.ElementsMatch(listA, listB)
	})

	t.Run("timeout", func(t *testing.T) {
		var list = make([]int, 0)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ctl := NewWorkerGroup(ctx, 2)
		ctl.Push(1, 3, 5, 7, 9)
		ctl.OnMessage = func(options interface{}) error {
			ctl.Lock()
			list = append(list, options.(int))
			ctl.Unlock()
			time.Sleep(2 * time.Second)
			return nil
		}
		ctl.StartAndWait()
		as.NoError(ctl.Err())
		as.ElementsMatch(list, []int{1, 3})
	})

	t.Run("empty", func(t *testing.T) {
		cc := NewWorkerGroup(context.Background(), 8)
		cc.OnMessage = func(options interface{}) error {
			return nil
		}
		cc.StartAndWait()
		as.NoError(cc.Err())
	})

	t.Run("one task", func(t *testing.T) {
		cc := NewWorkerGroup(context.Background(), 8)
		cc.Push(1)
		cc.OnMessage = func(options interface{}) error {
			return nil
		}
		cc.StartAndWait()
		as.NoError(cc.Err())
	})

	t.Run("error", func(t *testing.T) {
		cc := NewWorkerGroup(context.Background(), 8)
		cc.Push(1, 2)
		cc.OnMessage = func(options interface{}) error {
			var v = options.(int)
			if v%2 == 1 {
				return errors.New("test1")
			}
			return errors.New("test2")
		}
		cc.StartAndWait()
		as.Error(cc.Err())
	})

	t.Run("100 task", func(t *testing.T) {
		sum := int64(0)
		w := NewWorkerGroup(context.Background(), 8)
		for i := int64(1); i <= 100; i++ {
			w.Push(i)
		}
		w.OnMessage = func(options interface{}) error {
			atomic.AddInt64(&sum, options.(int64))
			return nil
		}
		w.StartAndWait()
		as.Equal(sum, int64(5050))
	})
}
