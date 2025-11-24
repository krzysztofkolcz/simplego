package exercises

func Counter() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

// func main() {
// 	c := counter()

// 	fmt.Println(c()) // 1
// 	fmt.Println(c()) // 2
// 	fmt.Println(c()) // 3
// }
