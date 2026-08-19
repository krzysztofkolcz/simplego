// Package block0 zawiera ćwiczenia z Bloku 0 (tablice, slice, mapy).
package block0

// Zadanie 2 — nil slice vs pusty slice
//
// var a []int   -> nil slice
// b := []int{}  -> pusty, ale nie nil
//
// describe(s []int) opisuje slice: czy jest nil, jaką ma len i cap.
// appendToNil() startuje od var a []int (nil) i robi append(a, 1) — pokazuje,
// że to bezpieczne i działa mimo że a jest nil.

import "fmt"

func describe(s []int) string {
	// TODO
	return fmt.Sprintf("nil=%v len=%d cap=%d", s == nil, len(s), cap(s))
}

func appendToNil() []int {
	var s []int
	return append(s, 1)
}
