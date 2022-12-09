package concurrency

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"sync"
)

type errorCollector struct {
	mu         *sync.RWMutex
	err        error
	succeedNum int
	failedNum  int
}

func (c *errorCollector) MarkSucceed() {
	c.mu.Lock()
	c.succeedNum++
	c.mu.Unlock()
}

func (c *errorCollector) MarkFailedWithError(err error) {
	c.mu.Lock()
	c.failedNum++
	c.err = multierror.Append(c.err, err)
	c.mu.Unlock()
}

func (c *errorCollector) MarkFailedWithMessage(msg string) {
	c.mu.Lock()
	c.failedNum++
	c.err = multierror.Append(c.err, errors.New(msg))
	c.mu.Unlock()
}

func (c *errorCollector) GetSucceedNum() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.succeedNum
}

func (c *errorCollector) GetFailedNum() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.failedNum
}

func (c *errorCollector) Err() error {
	c.mu.RLock()
	err := c.err
	c.mu.RUnlock()
	return err
}
