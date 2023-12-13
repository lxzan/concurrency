module github.com/lxzan/concurrency/benchmark

go 1.18

replace github.com/lxzan/concurrency v0.0.0 => ../

require (
	github.com/bytedance/gopkg v0.0.0-20230728082804-614d0af6619b
	github.com/lxzan/concurrency v0.0.0
	github.com/panjf2000/ants/v2 v2.8.1
)

require github.com/pkg/errors v0.9.1 // indirect
