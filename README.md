## Concurrency

[![Build Status](https://github.com/lxzan/concurrency/workflows/Go%20Test/badge.svg?branch=master)](https://github.com/lxzan/concurrency/actions?query=branch%3Amaster) [![Coverage Statusd][1]][2]

[1]: https://codecov.io/gh/lxzan/concurrency/branch/master/graph/badge.svg
[2]: https://codecov.io/gh/lxzan/concurrency

### install

```bash
GOPROXY=https://goproxy.cn go get -v github.com/lxzan/concurrency@latest
```

#### Usage

- WorkerGroup 任务组, 添加一组任务, 等待执行完成, 可以很好的替代`WaitGroup`.

```go
package main

import (
	"fmt"
	"github.com/lxzan/concurrency"
	"sync/atomic"
)

func main() {
	sum := int64(0)
	w := concurrency.NewWorkerGroup[int64]()
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

- WorkerQueue 任务队列, 可以不断往里面添加任务, 一旦有CPU资源空闲就去执行

```go
package main

import (
	"fmt"
	"github.com/lxzan/concurrency"
	"sync/atomic"
	"time"
)

func main() {
	sum := int64(0)
	w := concurrency.NewWorkerQueue()
	for i := int64(1); i <= 10; i++ {
		var x = i
		job := concurrency.FuncJob(func() {
			fmt.Printf("%v ", x)
			atomic.AddInt64(&sum, x)
		})
		w.Push(job)
	}
	w.Stop(time.Second)
	fmt.Printf("sum=%d\n", sum)
}
```

```
3 9 10 4 1 6 8 5 2 7 sum=55
```
