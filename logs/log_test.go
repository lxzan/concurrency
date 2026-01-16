package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	as := assert.New(t)

	t.Run("default logger", func(t *testing.T) {
		// 测试默认日志器不为空
		as.NotNil(DefaultLogger)
	})

	t.Run("logger interface", func(t *testing.T) {
		// 测试日志器实现了Logger接口
		var logger Logger = DefaultLogger
		as.NotNil(logger)
	})

	t.Run("errorf", func(t *testing.T) {
		// 测试Errorf方法不会panic
		logger := &logger{}
		as.NotPanics(func() {
			logger.Errorf("test error: %s", "message")
		})

		as.NotPanics(func() {
			logger.Errorf("test error: %d", 123)
		})

		as.NotPanics(func() {
			logger.Errorf("test error: %v", map[string]int{"key": 1})
		})
	})

	t.Run("multiple errorf calls", func(t *testing.T) {
		// 测试多次调用Errorf
		logger := &logger{}
		for i := 0; i < 10; i++ {
			as.NotPanics(func() {
				logger.Errorf("error %d", i)
			})
		}
	})

	t.Run("errorf with nil", func(t *testing.T) {
		// 测试Errorf处理nil值
		logger := &logger{}
		as.NotPanics(func() {
			logger.Errorf("error: %v", nil)
		})
	})

	t.Run("errorf with empty format", func(t *testing.T) {
		// 测试空格式字符串
		logger := &logger{}
		as.NotPanics(func() {
			logger.Errorf("")
		})
	})
}
