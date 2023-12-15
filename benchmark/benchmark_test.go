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

func newJob(wg *sync.WaitGroup) func() {
	return func() {
		fib(N)
		wg.Done()
	}
}

func Benchmark_Fib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(N)
	}
}

func Benchmark_StdGo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		job := newJob(wg)
		for j := 0; j < M; j++ {
			go job()
		}
		wg.Wait()
	}
}

func Benchmark_QueuesSingle(b *testing.B) {
	q := queues.New(
		queues.WithConcurrency(Concurrency),
		queues.WithSharding(1),
	)

	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		job := newJob(wg)
		for j := 0; j < M; j++ {
			q.Push(job)
		}
		wg.Wait()
	}
}

func Benchmark_QueuesMultiple(b *testing.B) {
	q := queues.New(
		queues.WithConcurrency(1),
		queues.WithSharding(Concurrency),
	)

	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		job := newJob(wg)
		for j := 0; j < M; j++ {
			q.Push(job)
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
		job := newJob(wg)
		for j := 0; j < M; j++ {
			q.Submit(job)
		}
		wg.Wait()
	}
}

func Benchmark_GoPool(b *testing.B) {
	q := gopool.NewPool("", Concurrency, gopool.NewConfig())

	for i := 0; i < b.N; i++ {
		wg := &sync.WaitGroup{}
		wg.Add(M)
		job := newJob(wg)
		for j := 0; j < M; j++ {
			q.Go(job)
		}
		wg.Wait()
	}
}
