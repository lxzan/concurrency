package queues

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lxzan/concurrency/logs"
	"github.com/stretchr/testify/assert"
)

func TestSingleQueue(t *testing.T) {
	as := assert.New(t)

	t.Run("sum", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(1000)
		w := New(WithConcurrency(16))
		for i := 1; i <= 1000; i++ {
			args := int64(i)
			w.Push(func() {
				atomic.AddInt64(&val, args)
				wg.Done()
			})
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		w := New(WithRecovery(), WithLogger(logs.DefaultLogger))
		w.Push(func() {
			panic("test")
		})
	})

	t.Run("stop timeout", func(t *testing.T) {
		cc := New(
			WithConcurrency(1),
			WithTimeout(100*time.Millisecond),
		)

		cc.Push(func() {
			time.Sleep(500 * time.Millisecond)
		})
		cc.Push(func() {
			time.Sleep(500 * time.Millisecond)
		})

		err := cc.Stop(context.Background())
		as.Error(err)
	})

	t.Run("stop graceful", func(t *testing.T) {
		cc := New(WithTimeout(time.Second))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop(context.Background())
		assert.Equal(t, int64(1), atomic.LoadInt64(&sum))
	})

	t.Run("", func(t *testing.T) {
		q := New(WithConcurrency(1))
		q.Push(func() { time.Sleep(100 * time.Millisecond) })
		q.Push(func() {})
		q.Stop(context.Background())
		q.Push(func() {})
		assert.Equal(t, 0, q.(*singleQueue).Len())
	})

	t.Run("stop", func(t *testing.T) {
		var q = New()
		var ctx = context.Background()
		q.Stop(ctx)
		q.Stop(ctx)
	})
}

func TestMultiQueue(t *testing.T) {
	as := assert.New(t)

	t.Run("sum", func(t *testing.T) {
		var val = int64(0)
		var wg = sync.WaitGroup{}
		wg.Add(1000)
		w := New(WithConcurrency(16), WithSharding(8))
		for i := 1; i <= 1000; i++ {
			args := int64(i)
			w.Push(func() {
				atomic.AddInt64(&val, args)
				wg.Done()
			})
		}
		wg.Wait()
		as.Equal(int64(500500), val)
	})

	t.Run("recover", func(t *testing.T) {
		w := New(WithRecovery(), WithLogger(logs.DefaultLogger), WithSharding(8))
		w.Push(func() {
			panic("test")
		})
	})

	t.Run("stop", func(t *testing.T) {
		cc := New(
			WithSharding(2),
			WithConcurrency(1),
		)
		as.Nil(cc.Stop(context.Background()))
	})

	t.Run("stop finished", func(t *testing.T) {
		cc := New(WithSharding(8))

		job := func() {
			time.Sleep(120 * time.Millisecond)
		}
		cc.Push(job)
		cc.Push(job)
		cc.Push(job)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(150 * time.Millisecond)
			cancel()
		}()
		t0 := time.Now()
		err := cc.Stop(ctx)
		println(time.Since(t0).String())
		as.Nil(err)
	})

	t.Run("stop timeout", func(t *testing.T) {
		cc := New(
			WithTimeout(100*time.Millisecond),
			WithSharding(2),
			WithConcurrency(1),
		)
		job := func() {
			time.Sleep(500 * time.Millisecond)
		}
		cc.Push(job)
		cc.Push(job)
		cc.Push(job)
		err := cc.Stop(context.Background())
		as.Error(err)
		as.Equal(cc.Len(), 1)
	})

	t.Run("stop graceful", func(t *testing.T) {
		cc := New(WithTimeout(time.Second), WithSharding(8))
		sum := int64(0)
		cc.Push(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&sum, 1)
		})
		cc.Stop(context.Background())
		assert.Equal(t, int64(1), atomic.LoadInt64(&sum))
	})

	t.Run("", func(t *testing.T) {
		q := New(WithConcurrency(1))
		q.Push(func() { time.Sleep(100 * time.Millisecond) })
		q.Push(func() {})
		q.Stop(context.Background())
		q.Push(func() {})
	})

	t.Run("push after stop", func(t *testing.T) {
		// 测试Stop后Push的行为
		q := New(WithConcurrency(2))
		var processed = int64(0)
		q.Push(func() {
			atomic.AddInt64(&processed, 1)
		})
		q.Stop(context.Background())
		as.Equal(int64(1), processed)

		// Stop后Push应该被忽略
		q.Push(func() {
			atomic.AddInt64(&processed, 1)
		})
		time.Sleep(50 * time.Millisecond)
		as.Equal(int64(1), processed) // 应该还是1，因为新任务被忽略
	})

	t.Run("concurrent push", func(t *testing.T) {
		// 测试并发Push
		q := New(WithConcurrency(4))
		var processed = int64(0)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				q.Push(func() {
					atomic.AddInt64(&processed, 1)
				})
			}()
		}
		wg.Wait()
		q.Stop(context.Background())
		as.Equal(int64(100), processed)
	})

	t.Run("zero concurrency", func(t *testing.T) {
		// 测试并发度为0的情况
		q := New(WithConcurrency(0))
		var processed = int64(0)
		q.Push(func() {
			atomic.AddInt64(&processed, 1)
		})
		q.Stop(context.Background())
		as.Equal(int64(1), processed)
	})

	t.Run("len accuracy", func(t *testing.T) {
		// 测试Len方法的准确性
		q := New(WithConcurrency(2))
		as.Equal(0, q.Len())
		q.Push(func() {})
		q.Push(func() {})
		// 由于有并发执行，Len可能为0或更小
		time.Sleep(10 * time.Millisecond)
		q.Stop(context.Background())
		as.Equal(0, q.Len())
	})

	t.Run("stop with empty queue", func(t *testing.T) {
		// 测试空队列Stop
		q := New()
		err := q.Stop(context.Background())
		as.NoError(err)
	})

	t.Run("stop multiple times", func(t *testing.T) {
		// 测试多次Stop
		q := New()
		err1 := q.Stop(context.Background())
		err2 := q.Stop(context.Background())
		as.NoError(err1)
		as.NoError(err2)
	})

	t.Run("context cancel during stop", func(t *testing.T) {
		// 测试Stop时context被取消
		q := New(WithConcurrency(1))
		q.Push(func() {
			time.Sleep(200 * time.Millisecond)
		})
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		err := q.Stop(ctx)
		as.Error(err)
	})

	t.Run("multiple queue error handling", func(t *testing.T) {
		// 测试多队列错误处理
		q := New(WithSharding(4), WithConcurrency(1), WithTimeout(50*time.Millisecond))
		// 添加一些会超时的任务
		for i := 0; i < 8; i++ {
			q.Push(func() {
				time.Sleep(200 * time.Millisecond)
			})
		}
		err := q.Stop(context.Background())
		// 应该返回超时错误
		as.Error(err)
	})

	t.Run("multiple queue all success", func(t *testing.T) {
		// 测试所有队列都成功的情况
		q := New(WithSharding(4), WithConcurrency(2))
		var processed = int64(0)
		for i := 0; i < 20; i++ {
			q.Push(func() {
				atomic.AddInt64(&processed, 1)
			})
		}
		err := q.Stop(context.Background())
		as.NoError(err)
		as.Equal(int64(20), processed)
	})
}
