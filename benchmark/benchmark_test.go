package benchmark

import (
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/lxzan/concurrency/queues"
	"github.com/panjf2000/ants/v2"
	"sync"
	"testing"
)

const (
	Concurrency = 8
	M           = 10000
	N           = 13
)

func Benchmark_Fib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(N)
	}
}

func Benchmark_StdGo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		for j := 0; j < M; j++ {
			go func() {
				fib(N)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func Benchmark_Queues(b *testing.B) {
	q := queues.New(queues.WithConcurrency(Concurrency))

	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		for j := 0; j < M; j++ {
			q.Push(func() {
				fib(N)
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func Benchmark_Ants(b *testing.B) {
	q, _ := ants.NewPool(Concurrency)
	defer q.Release()

	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		for j := 0; j < M; j++ {
			q.Submit(func() {
				fib(N)
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func Benchmark_GoPool(b *testing.B) {
	q := gopool.NewPool("", Concurrency, gopool.NewConfig())

	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		for j := 0; j < M; j++ {
			q.Go(func() {
				fib(N)
				wg.Done()
			})
		}
		wg.Wait()
	}
}
