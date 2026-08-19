package block0

import "fmt"

func verifySlices() {
	a := make([]int, 3, 6)
	s1 := a[0:2]
	s2 := a[0:6]
	fmt.Println(s1)
	fmt.Println(s2)
}

func demonstrateAppendPitfall() ([]int, []int, []int, []int) {
	base := []int{1, 2, 3, 4, 5}
	a := base[0:2]   // [1 2], len=2, cap=5
	c := base[0:2:2] // [1 2], len=2, cap=2
	b := base[2:4]   // [3 4], len=2, cap=3
	a = append(a, 99)
	c = append(c, 11)
	return base, a, b, c
}
