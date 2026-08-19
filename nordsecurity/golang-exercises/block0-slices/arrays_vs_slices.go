// Zadanie 1 — array vs slice
//
// Hipoteza: slice przekazuje wskaźnik na underlying array (w headerze), więc
// zmiana elementu w mutateSlice będzie widoczna u wywołującego. Array jest
// kopiowany w całości (wartość), więc mutateArray nie zmieni oryginału.
//
// TODO: zaimplementuj obie funkcje tak, żeby ustawiały element o indeksie 0
// na wartość 999.
package block0

import "errors"

func mutateArray(a [3]int) error {
	if len(a) < 1 {
		return errors.New("to short")
	}
	a[0] = 999
	return nil
}

func mutateSlice(s []int) error {
	if len(s) < 1 {
		return errors.New("to short")
	}
	s[0] = 999
	return nil
}
