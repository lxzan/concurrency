package benchmark

import (
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/lxzan/concurrency/queues"
	"github.com/panjf2000/ants/v2"
	"sync"
	"testing"
)

const (
	Concurrency = 16
	M           = 1000
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

func Benchmark_QueuesSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := queues.New(
			queues.WithConcurrency(Concurrency),
		)
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

func Benchmark_QueuesMultiple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := queues.New(
			queues.WithConcurrency(1),
			queues.WithSharding(Concurrency),
		)
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
	for i := 0; i < b.N; i++ {
		q, _ := ants.NewPool(Concurrency)
		wg := &sync.WaitGroup{}
		wg.Add(M)
		for j := 0; j < M; j++ {
			q.Submit(func() {
				fib(N)
				wg.Done()
			})
		}
		wg.Wait()
		q.Release()
	}
}

func Benchmark_GoPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := gopool.NewPool("", Concurrency, gopool.NewConfig())
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
