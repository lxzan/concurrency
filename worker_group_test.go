package concurrency

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTaskGroup(t *testing.T) {
	as := assert.New(t)

	t.Run("0 task", func(t *testing.T) {
		cc := NewWorkerGroup()
		err := cc.StartAndWait()
		as.NoError(err)
	})

	t.Run("1 task", func(t *testing.T) {
		cc := NewWorkerGroup()
		cc.AddJob(Job{
			Args: 1,
			Do: func(args interface{}) error {
				return nil
			},
		})
		err := cc.StartAndWait()
		as.NoError(err)
	})

	t.Run("100 task", func(t *testing.T) {
		sum := int64(0)
		w := NewWorkerGroup()
		for i := int64(1); i <= 100; i++ {
			w.AddJob(Job{
				Args: i,
				Do: func(args interface{}) error {
					atomic.AddInt64(&sum, args.(int64))
					return nil
				},
			})
		}
		_ = w.StartAndWait()
		as.Equal(sum, int64(5050))
	})

	t.Run("error", func(t *testing.T) {
		cc := NewWorkerGroup()
		cc.AddJob(
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
		err := cc.StartAndWait()
		as.Error(err)
	})

	t.Run("timeout", func(t *testing.T) {
		var mu = &sync.Mutex{}
		var list = make([]int, 0)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ctl := NewWorkerGroup(WithContext(ctx), WithConcurrency(2))
		var do = func(args interface{}) error {
			mu.Lock()
			list = append(list, args.(int))
			mu.Unlock()
			time.Sleep(2 * time.Second)
			return nil
		}
		ctl.AddJob(
			Job{1, do},
			Job{3, do},
			Job{5, do},
			Job{7, do},
			Job{9, do},
		)
		err := ctl.StartAndWait()
		as.NoError(err)
		as.ElementsMatch(list, []int{1, 3})
	})

	t.Run("recovery", func(t *testing.T) {
		ctl := NewWorkerGroup(WithRecovery())
		ctl.AddJob(Job{
			Args: nil,
			Do: func(args interface{}) error {
				return args.(error)
			},
		})
		err := ctl.StartAndWait()
		as.Error(err)
	})
}
