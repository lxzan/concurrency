## Concurrency

#### WorkerQueue 

> 工作队列, 可以不断往里面添加不同种类的任务, 一旦有资源空闲就去执行

```go
package main

import (
	"context"
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
	w := concurrency.NewWorkerQueue(context.Background(), 8)
	w.Push(
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


#### WorkerGroup 

> 工作组, 添加一组任务, 等待任务完全被执行

```go
package main

import (
	"context"
	"fmt"
	"github.com/lxzan/concurrency"
	"sync/atomic"
)

func main() {
	sum := int64(0)
	w := concurrency.NewWorkerGroup(context.Background(), 4)
	for i := int64(1); i <= 10; i++ {
		w.Push(concurrency.Job{
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