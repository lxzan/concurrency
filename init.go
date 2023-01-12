package concurrency

import (
	"github.com/panjf2000/ants/v2"
	"time"
)

var (
	gpool *ants.Pool
)

func DefaultPool() *ants.Pool {
	return gpool
}

func init() {
	p, _ := ants.NewPool(
		1024,
		ants.WithExpiryDuration(30*time.Second),
		ants.WithMaxBlockingTasks(64),
		ants.WithPanicHandler(nil),
	)
	gpool = p
}
