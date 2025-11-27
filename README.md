## Concurrency

[![Build Status](https://github.com/lxzan/concurrency/actions/workflows/go.yml/badge.svg)](https://github.com/lxzan/concurrency/actions/workflows/go.yml) [![Coverage Statusd][1]][2]

[1]: https://codecov.io/gh/lxzan/concurrency/branch/master/graph/badge.svg
[2]: https://codecov.io/gh/lxzan/concurrency

### Install

```bash
go get -v github.com/lxzan/concurrency@latest
```

#### Usage

##### 任务组
> 添加一组任务, 等待它们全部执行完成

```go
package main

import (
	"fmt"
	"github.com/lxzan/concurrency/groups"
	"sync/atomic"
)

func main() {
	sum := int64(0)
	w := groups.New[int64]()
	for i := int64(1); i <= 10; i++ {
		w.Push(i)
	}
	w.OnMessage = func(args int64) error {
		fmt.Printf("%v ", args)
		atomic.AddInt64(&sum, args)
		return nil
	}
	w.Start()
	fmt.Printf("sum=%d\n", sum)
}
```

```
4 5 6 7 8 9 10 1 3 2 sum=55
```

##### 任务队列
> 把任务加入队列, 异步执行

```go
package main

import (
	"context"
	"fmt"
	"github.com/lxzan/concurrency/queues"
	"sync/atomic"
)

func main() {
	sum := int64(0)
	w := queues.New()
	for i := int64(1); i <= 10; i++ {
		var x = i
		w.Push(func() {
			fmt.Printf("%v ", x)
			atomic.AddInt64(&sum, x)
		})
	}
	w.Stop(context.Background())
	fmt.Printf("sum=%d\n", sum)
}
```

```
3 9 10 4 1 6 8 5 2 7 sum=55
```

### Benchmark

```
go test -benchmem -run=^$ -bench . github.com/lxzan/concurrency/benchmark
goos: linux
goarch: amd64
pkg: github.com/lxzan/concurrency/benchmark
cpu: AMD Ryzen 5 PRO 4650G with Radeon Graphics
Benchmark_Fib-12                 1000000              1146 ns/op               0 B/op          0 allocs/op
Benchmark_StdGo-12                  3661            317905 ns/op           16064 B/op       1001 allocs/op
Benchmark_QueuesSingle-12           2178            532224 ns/op           67941 B/op       1098 allocs/op
Benchmark_QueuesMultiple-12         3691            317757 ns/op           61648 B/op       1256 allocs/op
Benchmark_Ants-12                   1569            751802 ns/op           22596 B/op       1097 allocs/op
Benchmark_GoPool-12                 2910            406935 ns/op           19042 B/op       1093 allocs/op
PASS
ok      github.com/lxzan/concurrency/benchmark  7.271s
```