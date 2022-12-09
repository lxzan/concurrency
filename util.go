package optimizer

import "context"

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func isCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
