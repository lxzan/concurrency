# Concurrency

[![Build Status](https://github.com/lxzan/concurrency/actions/workflows/go.yml/badge.svg)](https://github.com/lxzan/concurrency/actions/workflows/go.yml) [![Coverage Status][1]][2]

[1]: https://codecov.io/gh/lxzan/concurrency/branch/master/graph/badge.svg
[2]: https://codecov.io/gh/lxzan/concurrency

一个高性能的 Go 并发控制库，提供任务组和任务队列两种模式，支持灵活的并发控制和任务路由。

## 特性

- 🚀 **高性能**: 优化的并发控制机制，支持单队列和多队列模式
- 🔧 **灵活配置**: 支持自定义并发度、分片数、超时时间等参数
- 🎯 **任务路由**: 多队列模式支持自定义 hashcode 进行任务路由
- 🛡️ **安全可靠**: 完善的错误处理和恢复机制
- 📊 **高测试覆盖率**: 完整的单元测试，覆盖率达到 99%+

## 安装

```bash
go get -v github.com/lxzan/concurrency@latest
```

## 使用示例

### 任务组 (Groups)

任务组模式适用于需要等待所有任务执行完成的场景。添加一组任务后，调用 `Start()` 方法等待所有任务完成。

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
	
	// 添加任务
	for i := int64(1); i <= 10; i++ {
		w.Push(i)
	}
	
	// 设置任务处理函数
	w.OnMessage = func(args int64) error {
		fmt.Printf("%v ", args)
		atomic.AddInt64(&sum, args)
		return nil
	}
	
	// 启动并等待所有任务完成
	w.Start()
	fmt.Printf("\nsum=%d\n", sum)
}
```

输出示例：
```
4 5 6 7 8 9 10 1 3 2 
sum=55
```

### 任务队列 (Queues)

任务队列模式适用于异步执行任务的场景。任务会被加入队列并异步执行，调用 `Stop()` 方法等待所有任务完成。

#### 基础用法

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
	
	// 等待所有任务完成
	w.Stop(context.Background())
	fmt.Printf("\nsum=%d\n", sum)
}
```

输出示例：
```
3 9 10 4 1 6 8 5 2 7 
sum=55
```

#### 多队列模式与自定义路由

多队列模式通过分片降低锁竞争，提高并发性能。可以通过自定义 `hashcode` 参数控制任务路由到特定分片，实现任务的有序处理。

```go
package main

import (
	"context"
	"github.com/lxzan/concurrency/queues"
)

func main() {
	// 创建多队列，4个分片，每个分片并发度为2
	q := queues.New(
		queues.WithSharding(4),
		queues.WithConcurrency(2),
	)
	
	// 不指定hashcode，任务会轮询分配到各个分片
	q.Push(func() {
		// 任务1
	})
	
	// 指定hashcode，相同hashcode的任务会路由到同一个分片
	// 这对于需要保证同一用户/订单的任务有序处理很有用
	userID := int64(12345)
	q.Push(func() {
		// 处理用户12345的任务
	}, userID)
	
	q.Push(func() {
		// 用户12345的另一个任务，会路由到同一个分片
	}, userID)
	
	q.Stop(context.Background())
}
```

#### 配置选项

```go
q := queues.New(
	queues.WithSharding(8),              // 分片数（多队列模式）
	queues.WithConcurrency(16),          // 每个分片的并发度
	queues.WithTimeout(30*time.Second),  // 停止等待超时时间
	queues.WithRecovery(),                // 启用panic恢复
	queues.WithLogger(customLogger),      // 自定义日志记录器
)
```

## 性能基准测试

```
go test -benchmem -run=^$ -bench . github.com/lxzan/concurrency/benchmark
```

测试环境：Linux, AMD Ryzen 5 PRO 4650G

```
Benchmark_Fib-12                 1000000              1146 ns/op               0 B/op          0 allocs/op
Benchmark_StdGo-12                  3661            317905 ns/op           16064 B/op       1001 allocs/op
Benchmark_QueuesSingle-12           2178            532224 ns/op           67941 B/op       1098 allocs/op
Benchmark_QueuesMultiple-12         3691            317757 ns/op           61648 B/op       1256 allocs/op
Benchmark_Ants-12                   1569            751802 ns/op           22596 B/op       1097 allocs/op
Benchmark_GoPool-12                 2910            406935 ns/op           19042 B/op       1093 allocs/op
```

## 许可证

查看 [LICENSE](LICENSE) 文件了解详情。
