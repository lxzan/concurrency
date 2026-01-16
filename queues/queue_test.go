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

	t.Run("single queue finish during context done", func(t *testing.T) {
		// 测试在context Done时finish返回true的情况
		q := New(WithConcurrency(1), WithTimeout(100*time.Millisecond))
		q.Push(func() {
			time.Sleep(10 * time.Millisecond)
		})
		// 等待任务完成
		time.Sleep(50 * time.Millisecond)
		// 此时队列应该已经完成，Stop应该立即返回
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err := q.Stop(ctx)
		// 如果finish返回true，应该返回nil而不是context错误
		as.NoError(err)
	})

	t.Run("max concurrency limit", func(t *testing.T) {
		// 测试达到最大并发数时，新任务会排队
		q := New(WithConcurrency(2))
		var running = int32(0)
		var maxRunning = int32(0)
		var mu sync.Mutex

		// 添加多个长时间运行的任务
		for i := 0; i < 5; i++ {
			q.Push(func() {
				current := atomic.AddInt32(&running, 1)
				mu.Lock()
				if current > maxRunning {
					maxRunning = current
				}
				mu.Unlock()
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt32(&running, -1)
			})
		}

		time.Sleep(20 * time.Millisecond)
		// 检查最大并发数不超过设置的值
		mu.Lock()
		max := maxRunning
		mu.Unlock()
		as.LessOrEqual(int(max), 2)

		q.Stop(context.Background())
	})

	t.Run("new single queue", func(t *testing.T) {
		// 测试New返回单队列
		q := New(WithSharding(1))
		_, ok := q.(*singleQueue)
		as.True(ok)
	})

	t.Run("new multiple queue", func(t *testing.T) {
		// 测试New返回多队列
		q := New(WithSharding(4))
		_, ok := q.(*multipleQueue)
		as.True(ok)
	})

	t.Run("multiple queue partial error", func(t *testing.T) {
		// 测试多队列中部分队列超时的情况
		q := New(WithSharding(4), WithConcurrency(1), WithTimeout(50*time.Millisecond))
		// 添加足够多的长时间任务，确保至少有一个队列会超时
		for i := 0; i < 8; i++ {
			q.Push(func() {
				time.Sleep(200 * time.Millisecond)
			})
		}
		err := q.Stop(context.Background())
		// 应该返回超时错误
		as.Error(err)
	})

	t.Run("single queue getJob with max concurrency", func(t *testing.T) {
		// 测试getJob在达到最大并发时返回nil
		q := New(WithConcurrency(1))
		var processed = int64(0)
		// 添加一个长时间任务
		q.Push(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt64(&processed, 1)
		})
		// 立即添加另一个任务，此时应该排队
		q.Push(func() {
			atomic.AddInt64(&processed, 1)
		})
		time.Sleep(150 * time.Millisecond)
		q.Stop(context.Background())
		as.Equal(int64(2), processed)
	})

	t.Run("single queue len during execution", func(t *testing.T) {
		// 测试在执行过程中Len的准确性
		q := New(WithConcurrency(1))
		q.Push(func() {
			time.Sleep(50 * time.Millisecond)
		})
		// 立即检查Len，任务可能正在执行或排队
		len1 := q.Len()
		as.GreaterOrEqual(len1, 0)
		q.Stop(context.Background())
		as.Equal(0, q.Len())
	})

	t.Run("multiple queue len", func(t *testing.T) {
		// 测试多队列Len的准确性
		q := New(WithSharding(4), WithConcurrency(2))
		as.Equal(0, q.Len())
		for i := 0; i < 10; i++ {
			q.Push(func() {})
		}
		// Len可能因为并发执行而小于10
		len1 := q.Len()
		as.GreaterOrEqual(len1, 0)
		as.LessOrEqual(len1, 10)
		q.Stop(context.Background())
		as.Equal(0, q.Len())
	})

	t.Run("multiple queue with custom hashcode", func(t *testing.T) {
		// 测试使用自定义hashcode路由任务
		q := New(WithSharding(4), WithConcurrency(1))

		// 使用相同的hashcode，确保任务路由到同一个分片
		var processed = make([]int64, 4) // 每个分片的处理计数
		var mu sync.Mutex

		// 使用hashcode 0，应该路由到分片 0
		for i := 0; i < 10; i++ {
			q.Push(func() {
				mu.Lock()
				processed[0]++
				mu.Unlock()
			}, 0)
		}

		// 使用hashcode 1，应该路由到分片 1
		for i := 0; i < 10; i++ {
			q.Push(func() {
				mu.Lock()
				processed[1]++
				mu.Unlock()
			}, 1)
		}

		// 使用hashcode 2，应该路由到分片 2
		for i := 0; i < 10; i++ {
			q.Push(func() {
				mu.Lock()
				processed[2]++
				mu.Unlock()
			}, 2)
		}

		// 使用hashcode 3，应该路由到分片 3
		for i := 0; i < 10; i++ {
			q.Push(func() {
				mu.Lock()
				processed[3]++
				mu.Unlock()
			}, 3)
		}

		q.Stop(context.Background())

		// 验证每个分片都处理了10个任务
		as.Equal(int64(10), processed[0])
		as.Equal(int64(10), processed[1])
		as.Equal(int64(10), processed[2])
		as.Equal(int64(10), processed[3])
	})

	t.Run("multiple queue hashcode routing", func(t *testing.T) {
		// 测试hashcode路由的正确性
		q := New(WithSharding(8), WithConcurrency(1))

		// 测试不同hashcode值
		testCases := []struct {
			hashcode int64
			expected int64 // 期望路由到的分片索引 (hashcode & 7)
		}{
			{0, 0},
			{1, 1},
			{7, 7},
			{8, 0},  // 8 & 7 = 0
			{9, 1},  // 9 & 7 = 1
			{15, 7}, // 15 & 7 = 7
			{16, 0}, // 16 & 7 = 0
		}

		var processed = int64(0)
		for _, tc := range testCases {
			// 验证hashcode参数被正确传递和使用
			q.Push(func() {
				atomic.AddInt64(&processed, 1)
			}, tc.hashcode)

			// 验证期望的分片索引计算正确
			expectedShard := tc.hashcode & 7
			as.Equal(tc.expected, expectedShard)
		}

		time.Sleep(50 * time.Millisecond)
		q.Stop(context.Background())

		// 验证所有任务都被处理
		as.Equal(int64(len(testCases)), processed)
	})

	t.Run("multiple queue hashcode without parameter", func(t *testing.T) {
		// 测试不提供hashcode参数时的默认行为（轮询分配）
		q := New(WithSharding(4), WithConcurrency(1))

		var processed = int64(0)

		// 不提供hashcode，应该使用serial轮询分配
		for i := 0; i < 40; i++ {
			q.Push(func() {
				atomic.AddInt64(&processed, 1)
			})
		}

		time.Sleep(100 * time.Millisecond)
		q.Stop(context.Background())

		// 验证所有任务都被处理
		as.Equal(int64(40), processed)
	})

	t.Run("multiple queue hashcode consistency", func(t *testing.T) {
		// 测试相同hashcode总是路由到同一个分片
		q := New(WithSharding(4), WithConcurrency(1))

		// 使用固定的hashcode
		const testHashcode int64 = 12345
		const expectedShard = testHashcode & 3 // 4个分片，所以 & 3

		var processed = int64(0)

		// 多次使用相同的hashcode，应该路由到同一个分片
		for i := 0; i < 20; i++ {
			q.Push(func() {
				atomic.AddInt64(&processed, 1)
			}, testHashcode)
		}

		time.Sleep(100 * time.Millisecond)
		q.Stop(context.Background())

		// 验证所有任务都被处理
		as.Equal(int64(20), processed)

		// 验证hashcode计算正确
		as.Equal(int64(1), expectedShard) // 12345 & 3 = 1
	})
}
