package benchmark

func fib(n int) int {
	switch n {
	case 0, 1:
		return n
	default:
		return fib(n-1) + fib(n-2)
	}
}
