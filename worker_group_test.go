package concurrency

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTaskGroup(t *testing.T) {
	as := assert.New(t)
	t.Run("success", func(t *testing.T) {
		listA := make([]uint8, 0)
		listB := make([]uint8, 0)
		ctl := NewWorkerGroup(context.Background(), 8)
		for i := 0; i < 100; i++ {
			ctl.Push(Job{
				Args: uint8(i),
				Do: func(args interface{}) error {
					ctl.Lock()
					listA = append(listA, args.(uint8))
					ctl.Unlock()
					return nil
				},
			})
			listB = append(listB, uint8(i))
		}
		ctl.StartAndWait()
		as.ElementsMatch(listA, listB)
	})

	t.Run("timeout", func(t *testing.T) {
		var list = make([]int, 0)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ctl := NewWorkerGroup(ctx, 2)
		var do = func(args interface{}) error {
			ctl.Lock()
			list = append(list, args.(int))
			ctl.Unlock()
			time.Sleep(2 * time.Second)
			return nil
		}
		ctl.Push(
			Job{1, do},
			Job{3, do},
			Job{5, do},
			Job{7, do},
			Job{9, do},
		)
		ctl.StartAndWait()
		as.NoError(ctl.Err())
		as.ElementsMatch(list, []int{1, 3})
	})

	t.Run("empty", func(t *testing.T) {
		cc := NewWorkerGroup(context.Background(), 8)
		cc.StartAndWait()
		as.NoError(cc.Err())
	})

	t.Run("one task", func(t *testing.T) {
		cc := NewWorkerGroup(context.Background(), 8)
		cc.Push(Job{
			Args: 1,
			Do: func(args interface{}) error {
				return nil
			},
		})
		cc.StartAndWait()
		as.NoError(cc.Err())
	})

	t.Run("error", func(t *testing.T) {
		cc := NewWorkerGroup(context.Background(), 8)
		cc.Push(
			Job{
				Args: 1,
				Do: func(args interface{}) error {
					return errors.New("test1")
				},
			},
			Job{
				Args: 2,
				Do: func(args interface{}) error {
					return errors.New("test2")
				},
			},
		)
		cc.StartAndWait()
		as.Error(cc.Err())
	})

	t.Run("100 task", func(t *testing.T) {
		sum := int64(0)
		w := NewWorkerGroup(context.Background(), 8)
		for i := int64(1); i <= 100; i++ {
			w.Push(Job{
				Args: i,
				Do: func(args interface{}) error {
					atomic.AddInt64(&sum, args.(int64))
					return nil
				},
			})
		}
		w.StartAndWait()
		as.Equal(sum, int64(5050))
	})
}
