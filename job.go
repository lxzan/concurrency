package concurrency

type (
	// 任务抽象
	Job interface {
		Do()
	}

	// 无参数的快捷任务
	FuncJob func()
)

type parameterizedJob struct {
	args interface{}
	do   func(args interface{})
}

func (f FuncJob) Do() {
	f()
}

// 带参数任务
func ParameterizedJob(args interface{}, f func(args interface{})) Job {
	return &parameterizedJob{args: args, do: f}
}

func (c *parameterizedJob) Do() { c.do(c.args) }
