package lesson001

import "fmt"

func Test(){
	ch := make(chan int)

	go func() {
		ch <- 42 // wysyłanie (blokujące)
	}()

	val := <-ch // odbieranie (blokujące)
	fmt.Println(val)
}