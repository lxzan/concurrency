package concurrency

type (
	// 任务抽象
	Job interface {
		Do() error
	}

	// 无参数的快捷任务
	QuickJob func() error
)

type parameterizedJob struct {
	args []interface{}
	do   func(args ...interface{}) error
}

func (f QuickJob) Do() error {
	return f()
}

// 带参数任务
func ParameterizedJob(f func(args ...interface{}) error, args ...interface{}) Job {
	return &parameterizedJob{args: args, do: f}
}

func (c *parameterizedJob) Do() error { return c.do(c.args...) }
