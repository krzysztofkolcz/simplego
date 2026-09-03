package block2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFanOutFanIn(t *testing.T) {
	t.Run("something", func(t *testing.T) {
		expected := []int{}
		for i := 0; i < 100; i++ {
			expected = append(expected, i)
		}
		input := make(chan int)
		go func() {
			for i := 0; i < 100; i++ {
				input <- i
			}
			close(input)
		}()
		output := FanOutFanIn(input, 10, Process)
		var results []int
		for v := range output {
			results = append(results, v)
		}
		assert.ElementsMatch(t, expected, results)
	})
}
