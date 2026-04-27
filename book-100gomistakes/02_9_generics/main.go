package main

import (
	"fmt"
	"simpleGo/book-100gomistakes/02_9_generics/examples"
	"sort"
)

func main() {
	s := examples.SliceFn[int]{
		S: []int{3, 2, 1},
		Compare: func(a, b int) bool {
			return a < b
		},
	}
	sort.Sort(s)
	fmt.Println(s.S)
}
