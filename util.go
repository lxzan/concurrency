package concurrency

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

func recoveryCaller[T any](args T, f func(T) error) (err error) {
	defer func() {
		if fatalError := recover(); fatalError != nil {
			var msg = make([]byte, 0, 256)
			msg = append(msg, fmt.Sprintf("fatal error: %v\n", fatalError)...)
			for i := 1; true; i++ {
				_, caller, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				if !strings.Contains(caller, "src/runtime") {
					msg = append(msg, fmt.Sprintf("caller: %s, line: %d\n", caller, line)...)
				}
			}
			err = errors.New(string(msg))
		}
	}()
	return f(args)
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
