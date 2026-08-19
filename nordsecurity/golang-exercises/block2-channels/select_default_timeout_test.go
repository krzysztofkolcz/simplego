package block2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTryReceiveNonBlocking(t *testing.T) {
	ch := make(chan int)
	ch2 := make(chan int, 1)
	t.Run("unbuffered nothing in channel", func(t *testing.T) {
		value, ok := TryReceiveNonBlocking(ch)
		assert.Equal(t, 0, value)
		assert.False(t, ok)
	})
	t.Run("unbuffered something in channel", func(t *testing.T) {
		ch2 <- 1
		value, ok := TryReceiveNonBlocking(ch2)
		assert.Equal(t, 1, value)
		assert.True(t, ok)
	})
}

func TestReceiveWithTimeout(t *testing.T) {
	ch := make(chan int)
	t.Run("timeout", func(t *testing.T) {
		value, ok := ReceiveWithTimeout(ch, 1*time.Second)
		assert.Equal(t, 0, value)
		assert.True(t, ok)
	})

	ch2 := make(chan int, 1)
	t.Run("value in channel", func(t *testing.T) {
		ch2 <- 17
		value, ok := ReceiveWithTimeout(ch2, 1*time.Second)
		assert.Equal(t, 17, value)
		assert.False(t, ok)
	})

	t.Run("send before timeout", func(t *testing.T) {
		go func(ch chan int) {
			time.Sleep(100 * time.Millisecond)
			ch <- 17
		}(ch)
		value, ok := ReceiveWithTimeout(ch, 1*time.Second)
		assert.Equal(t, 17, value)
		assert.False(t, ok)
	})
}
