package lesson003

import "fmt"

func TestSwitch(){
	ch1 := make(chan int)
	ch2 := make(chan int)


	go func() {
		ch1 <- 1
		fmt.Println("sent to ch1")
	}()

	go func() {
		ch2 <- 1
		fmt.Println("sent to ch2")
	}()

	go func(){
		select {
		case msg := <-ch1:
			fmt.Println("ch1:", msg)
		case msg := <-ch2:
			fmt.Println("ch2:", msg)
		}
	}()

	// select {
	// case msg := <-ch1:
	// 	fmt.Println("ch1:", msg)
	// case msg := <-ch2:
	// 	fmt.Println("ch2:", msg)
	// }
}