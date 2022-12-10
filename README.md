## Concurrency

[![Build Status](https://github.com/lxzan/concurrency/workflows/Go%20Test/badge.svg?branch=master)](https://github.com/lxzan/concurrency/actions?query=branch%3Amaster)

### install
```bash
GOPROXY=https://goproxy.cn go get -v github.com/lxzan/concurrency@latest
```

#### Feature
- 最大并发协程数量限制
- 支持 `contex.Contex`
- 支持 `panic recover`, 返回包含错误堆栈的 `error`
- 任务调度不依赖 `time.Ticker` 和 `channel`


#### Usage
- WorkerQueue  工作队列, 可以不断往里面添加任务, 一旦有CPU资源空闲就去执行

```go
package main

import (
	"fmt"
	"github.com/lxzan/concurrency"
	"time"
)

func Add(args interface{}) error {
	arr := args.([]int)
	ans := 0
	for _, item := range arr {
		ans += item
	}
	fmt.Printf("args=%v, ans=%d\n", args, ans)
	return nil
}

func Mul(args interface{}) error {
	arr := args.([]int)
	ans := 1
	for _, item := range arr {
		ans *= item
	}
	fmt.Printf("args=%v, ans=%d\n", args, ans)
	return nil
}

func main() {
	args1 := []int{1, 3}
	args2 := []int{1, 3, 5}
	w := concurrency.NewWorkerQueue()
	w.AddJob(
		concurrency.Job{Args: args1, Do: Add},
		concurrency.Job{Args: args1, Do: Mul},
		concurrency.Job{Args: args2, Do: Add},
		concurrency.Job{Args: args2, Do: Mul},
	)
	w.StopAndWait(30*time.Second)
}
```

```
args=[1 3], ans=4
args=[1 3 5], ans=15
args=[1 3], ans=3
args=[1 3 5], ans=9
```


- WorkerGroup 工作组, 添加一组任务, 等待执行完成, 可以很好的替代`WaitGroup`.

```go
package main

import (
	"fmt"
	"github.com/lxzan/concurrency"
	"sync/atomic"
)

func main() {
	sum := int64(0)
	w := concurrency.NewWorkerGroup()
	for i := int64(1); i <= 10; i++ {
		w.AddJob(concurrency.Job{
			Args: i,
			Do: func(args interface{}) error {
				fmt.Printf("%v ", args)
				atomic.AddInt64(&sum, args.(int64))
				return nil
			},
		})
	}
	w.StartAndWait()
	fmt.Printf("sum=%d\n", sum)
}
```

```
4 5 6 7 8 9 10 1 3 2 sum=55
```
