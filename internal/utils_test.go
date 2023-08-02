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
