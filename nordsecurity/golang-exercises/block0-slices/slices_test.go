package block0

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySlices(t *testing.T) {
	t.Run("something", func(t *testing.T) {
		verifySlices()
		assert.True(t, true)
	})
}

func TestDemonstratePitfall(t *testing.T) {
	expBase := []int{1, 2, 99, 4, 5}
	expA := []int{1, 2, 99}
	expB := []int{99, 4}
	expC := []int{1, 2, 11}
	t.Run("something", func(t *testing.T) {
		base, a, b, c := demonstrateAppendPitfall()
		assert.Equal(t, expBase, base)
		assert.Equal(t, expA, a)
		assert.Equal(t, expB, b)
		assert.Equal(t, expC, c)
	})
}
