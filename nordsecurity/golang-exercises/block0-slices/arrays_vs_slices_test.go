package block0

import "testing"

func TestMutateArray_NieZmieniaOryginalu(t *testing.T) {
	original := [3]int{1, 2, 3}

	if err := mutateArray(original); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if original[0] == 999 {
		t.Fatalf("mutateArray zmienił oryginał — array powinien być kopiowany przez wartość, original=%v", original)
	}
}

func TestMutateSlice_ZmieniaOryginal(t *testing.T) {
	original := []int{1, 2, 3}

	if err := mutateSlice(original); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if original[0] != 999 {
		t.Fatalf("mutateSlice nie zmienił oryginału — oczekiwano 999, original=%v", original)
	}
}
