package examples

import "testing"

func TestGetKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[K]V
		want []K
	}{
		{
			name: "mapping",
			m : map[CustomInt]int{
				1: 1,
				2: 2,
				3: 3,
			},
			want: [1,2,3],
		}
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetKeys(tt.m)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
