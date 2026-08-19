package block2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProduceAndRange(t *testing.T) {
	expected := []int{1, 2, 3}
	t.Run("something", func(t *testing.T) {
		result := ProduceAndRange()
		assert.Equal(t, expected, result)
	})
}

func TestSendToClosedChannel(t *testing.T) {
	t.Run("something", func(t *testing.T) {
		panicked, msg := SendToClosedChannel()
		assert.True(t, panicked)
		assert.Equal(t, "send on closed channel", msg)
	})
}
