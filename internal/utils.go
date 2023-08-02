package internal

type Integer interface {
	int | int64 | int32 | uint | uint64 | uint32
}

func Min[T Integer](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func ToBinaryNumber[T Integer](n T) T {
	var x T = 1
	for x < n {
		x *= 2
	}
	return x
}
