package concurrency

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestNewTaskGroup(t *testing.T) {
	as := assert.New(t)

	t.Run("0 task", func(t *testing.T) {
		cc := NewWorkerGroup[int]()
		err := cc.Start()
		as.NoError(err)
	})

	t.Run("1 task", func(t *testing.T) {
		cc := NewWorkerGroup[int]()
		cc.Push(0)
		err := cc.Start()
		as.NoError(err)
	})

	t.Run("100 task", func(t *testing.T) {
		sum := int64(0)
		w := NewWorkerGroup[int64]()
		w.OnMessage = func(args int64) error {
			atomic.AddInt64(&sum, args)
			return nil
		}
		for i := int64(1); i <= 100; i++ {
			w.Push(i)
		}
		_ = w.Start()
		as.Equal(sum, int64(5050))
	})

	t.Run("error", func(t *testing.T) {
		cc := NewWorkerGroup[int]()
		cc.Push(1)
		cc.Push(2)
		cc.OnMessage = func(args int) error {
			return errors.New("test1")
		}
		err := cc.Start()
		as.Error(err)
	})

	//t.Run("timeout", func(t *testing.T) {
	//	var mu = &sync.Mutex{}
	//	var list = make([]int, 0)
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	defer cancel()
	//	ctl := NewWorkerGroup[int]().WithContext(ctx).WithConcurrency(2)
	//	ctl.Push(1, 3, 5, 7, 9)
	//	ctl.OnMessage = func(args int) error {
	//		mu.Lock()
	//		list = append(list, args)
	//		mu.Unlock()
	//		time.Sleep(2 * time.Second)
	//		return nil
	//	}
	//	err := ctl.Start()
	//	as.NoError(err)
	//	as.ElementsMatch(list, []int{1, 3})
	//})

	t.Run("recovery", func(t *testing.T) {
		ctl := NewWorkerGroup[int]()
		ctl.Push(1)
		ctl.Push(2)
		ctl.OnMessage = func(args int) error {
			var err error
			println(err.Error())
			return err
		}
		err := ctl.Start()
		as.Error(err)
	})
}
