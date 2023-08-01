package benchmark

const (
	Concurrency = 8
	M           = 1000
	N           = 1000
)

func f() int {
	sum := 0
	for i := 0; i < N; i++ {
		sum += i
	}
	return sum
}
