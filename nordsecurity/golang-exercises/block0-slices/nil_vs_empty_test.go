package block0

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  string
	}{
		{
			name:  "2 elements",
			input: []int{19, 29},
			want:  "nil=false len=2 cap=2",
		},
		{
			name:  "nil",
			input: nil,
			want:  "nil=true len=0 cap=0",
		},
		{
			name:  "empty",
			input: []int{},
			want:  "nil=false len=0 cap=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := describe(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestAppendToNil(t *testing.T) {
	t.Run("something", func(t *testing.T) {
		result := appendToNil()
		s := []int{1}
		assert.Equal(t, s, result)
	})
}
