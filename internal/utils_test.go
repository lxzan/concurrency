package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToBinaryNumber(t *testing.T) {
	assert.Equal(t, 8, ToBinaryNumber(7))
	assert.Equal(t, 1, ToBinaryNumber(0))
	assert.Equal(t, 128, ToBinaryNumber(120))
	assert.Equal(t, 1024, ToBinaryNumber(1024))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1, Min(1, 2))
	assert.Equal(t, 3, Min(4, 3))
}

func TestIsSameSlice(t *testing.T) {
	assert.True(t, IsSameSlice(
		[]int{1, 2, 3},
		[]int{1, 2, 3},
	))

	assert.False(t, IsSameSlice(
		[]int{1, 2, 3},
		[]int{1, 2},
	))

	assert.False(t, IsSameSlice(
		[]int{1, 2, 3},
		[]int{1, 2, 4},
	))
}

func TestSelectValue(t *testing.T) {
	assert.Equal(t, SelectValue(true, 1, 2), 1)
	assert.Equal(t, SelectValue(false, 1, 2), 2)
}
