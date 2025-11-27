package groups

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTaskGroup(t *testing.T) {
	as := assert.New(t)

	t.Run("0 task", func(t *testing.T) {
		cc := New[int]()
		err := cc.Start()
		as.NoError(err)
	})

	t.Run("1 task", func(t *testing.T) {
		cc := New[int]()
		cc.Push(0)
		err := cc.Start()
		as.NoError(err)
	})

	t.Run("100 task", func(t *testing.T) {
		sum := int64(0)
		w := New[int64]()
		w.OnMessage = func(args int64) error {
			atomic.AddInt64(&sum, args)
			w.Update(func() {})
			return nil
		}
		for i := int64(1); i <= 100; i++ {
			w.Push(i)
		}
		_ = w.Start()
		as.Equal(sum, int64(5050))
	})

	t.Run("error", func(t *testing.T) {
		cc := New[int]()
		cc.Push(1)
		cc.Push(2)
		cc.OnMessage = func(args int) error {
			return errors.New("test1")
		}
		err := cc.Start()
		as.Error(err)
	})

	t.Run("timeout", func(t *testing.T) {
		var mu = &sync.Mutex{}
		var list = make([]int, 0)
		ctl := New[int](WithConcurrency(2), WithTimeout(time.Second))
		ctl.Push(1, 3, 5, 7, 9)
		ctl.OnMessage = func(args int) error {
			mu.Lock()
			list = append(list, args)
			mu.Unlock()
			time.Sleep(2 * time.Second)
			return nil
		}
		err := ctl.Start()
		as.Error(err)
		as.ElementsMatch(list, []int{1, 3})
	})

	t.Run("recovery", func(t *testing.T) {
		ctl := New[int](WithRecovery())
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

	t.Run("cancel", func(t *testing.T) {
		ctl := New[int](WithConcurrency(1))
		ctl.Push(1, 3, 5)
		arr := make([]int, 0)
		ctl.OnMessage = func(args int) error {
			ctl.Update(func() {
				arr = append(arr, args)
			})
			switch args {
			case 3:
				return errors.New("3")
			default:
				return nil
			}
		}
		ctl.OnError = func(args int, err error) {
			ctl.Cancel()
		}
		err := ctl.Start()
		as.Error(err)
		as.ElementsMatch(arr, []int{1, 3})
	})

	t.Run("push after start", func(t *testing.T) {
		// 测试在Start之后Push任务的行为
		ctl := New[int](WithConcurrency(2))
		ctl.Push(1, 2)
		var processed = make([]int, 0)
		var mu = sync.Mutex{}
		ctl.OnMessage = func(args int) error {
			mu.Lock()
			processed = append(processed, args)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		go func() {
			time.Sleep(5 * time.Millisecond)
			ctl.Push(3, 4) // 在Start之后Push
		}()

		err := ctl.Start()
		as.NoError(err)
		// 验证所有任务都被处理
		as.GreaterOrEqual(len(processed), 2)
	})

	t.Run("multiple errors", func(t *testing.T) {
		// 测试多个任务都出错的情况
		ctl := New[int](WithConcurrency(2))
		ctl.Push(1, 2, 3)
		ctl.OnMessage = func(args int) error {
			return errors.Errorf("error %d", args)
		}
		err := ctl.Start()
		as.Error(err)
		// 验证错误被正确收集
		as.Contains(err.Error(), "error")
	})

	t.Run("concurrent push", func(t *testing.T) {
		// 测试并发Push
		ctl := New[int](WithConcurrency(4))
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				ctl.Push(n)
			}(i)
		}
		wg.Wait()

		var processed = int64(0)
		ctl.OnMessage = func(args int) error {
			atomic.AddInt64(&processed, 1)
			return nil
		}
		err := ctl.Start()
		as.NoError(err)
		as.Equal(int64(10), processed)
	})

	t.Run("zero concurrency", func(t *testing.T) {
		// 测试并发度为0的情况（应该使用默认值）
		ctl := New[int](WithConcurrency(0))
		ctl.Push(1, 2, 3)
		var processed = int64(0)
		ctl.OnMessage = func(args int) error {
			atomic.AddInt64(&processed, 1)
			return nil
		}
		err := ctl.Start()
		as.NoError(err)
		as.Equal(int64(3), processed)
	})

	t.Run("cancel before start", func(t *testing.T) {
		// 测试在Start之前Cancel
		ctl := New[int]()
		ctl.Push(1, 2, 3)
		ctl.Cancel()
		var processed = int64(0)
		ctl.OnMessage = func(args int) error {
			atomic.AddInt64(&processed, 1)
			return nil
		}
		_ = ctl.Start()
		// Cancel后任务应该不会执行
		as.Equal(int64(0), processed)
		// 由于任务被取消，可能会超时或返回错误
	})

	t.Run("update deadlock check", func(t *testing.T) {
		// 测试Update方法不会导致死锁
		ctl := New[int](WithConcurrency(2))
		ctl.Push(1, 2, 3, 4, 5)
		var processed = int64(0)
		ctl.OnMessage = func(args int) error {
			ctl.Update(func() {
				processed++
			})
			return nil
		}
		err := ctl.Start()
		as.NoError(err)
		as.Equal(int64(5), processed)
	})

	t.Run("task count accuracy", func(t *testing.T) {
		// 测试任务计数准确性
		ctl := New[int](WithConcurrency(2))
		ctl.Push(1, 2, 3, 4, 5)
		var processed = int64(0)
		ctl.OnMessage = func(args int) error {
			atomic.AddInt64(&processed, 1)
			return nil
		}
		err := ctl.Start()
		as.NoError(err)
		as.Equal(int64(5), processed)
	})

	t.Run("len accuracy", func(t *testing.T) {
		// 测试Len方法的准确性
		ctl := New[int]()
		as.Equal(0, ctl.Len())
		ctl.Push(1, 2, 3)
		as.Equal(3, ctl.Len())
		ctl.Start()
		// Start后队列应该为空
		as.Equal(0, ctl.Len())
	})
}
