## Concurrency

[![Build Status](https://github.com/lxzan/concurrency/workflows/Go%20Test/badge.svg?branch=master)](https://github.com/lxzan/concurrency/actions?query=branch%3Amaster) [![Coverage Statusd][1]][2]

[1]: https://codecov.io/gh/lxzan/concurrency/branch/master/graph/badge.svg
[2]: https://codecov.io/gh/lxzan/concurrency

### Install

```bash
GOPROXY=https://goproxy.cn go get -v github.com/lxzan/concurrency@latest
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
	w.Stop()
	fmt.Printf("sum=%d\n", sum)
}
```

```
3 9 10 4 1 6 8 5 2 7 sum=55
```

### Benchmark

```
go test -benchmem -run=^$ -bench . github.com/lxzan/concurrency/benchmark
goos: darwin
goarch: arm64
pkg: github.com/lxzan/concurrency/benchmark
Benchmark_Fib-8          1485490               775.0 ns/op             0 B/op          0 allocs/op
Benchmark_StdGo-8            388           3066459 ns/op          160537 B/op      10002 allocs/op
Benchmark_Queues-8           457           2602319 ns/op          324489 B/op      11061 allocs/op
Benchmark_Ants-8             139           7337507 ns/op          160368 B/op      10004 allocs/op
Benchmark_GoPool-8           264           4514672 ns/op          191897 B/op      10569 allocs/op
PASS
ok      github.com/lxzan/concurrency/benchmark  8.500s
```