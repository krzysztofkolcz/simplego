package exercises_test

import (
	"simpleGo/100gomistakes/011_functional_options_pattern/exercises"
	"testing"
)

func TestCounter(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want func() int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercises.Counter()
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Counter() = %v, want %v", got, tt.want)
			}
		})
	}
}
