package benchmark

import (
	"github.com/lxzan/concurrency/queues"
	"github.com/panjf2000/ants/v2"
	"sync"
	"testing"
)

func Benchmark_Queues(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := queues.New(queues.WithConcurrency(Concurrency))
		wg := &sync.WaitGroup{}
		wg.Add(M)
		for j := 0; j < M; j++ {
			q.Push(queues.FuncJob(func() {
				f()
				wg.Done()
			}))
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
				f()
				wg.Done()
			})
		}
		wg.Wait()
		q.Release()
	}
}
