package lesson002

import (
	"fmt"
	"time"
)

func TestUnbufferedBlocking(){
	ch := make(chan int)

	go func() {
		ch <- 1
		fmt.Println("sent") // sent pojawi się dopiero przy odbiorze
	}()

	time.Sleep(time.Second)
	<-ch
}

// w ogóle nic nie wypisuje - czemu?
func TestUnbufferedBlocking2(){
	ch := make(chan int)


    for i := 0; i < 5; i++ {
		go func() {
			ch <- 1
			fmt.Println("sent") // sent pojawi się dopiero przy odbiorze
		}()
    }

	time.Sleep(time.Second)
	<-ch // ale odbieram tylko raz, więc się pewnie zatnie
}


/*
	wysyłanie blokuje tylko gdy bufor pełny
	odbieranie blokuje gdy pusty
*/
func TestBuffered(){
	ch := make(chan int, 1)

	go func() {
		ch <- 1
		fmt.Println("sent") // sent pojawi się od razu
	}()

	time.Sleep(time.Second)
	<-ch
}


func TestBuffered2(){
	ch := make(chan int, 1)

	go func() {
		ch <- 1
		fmt.Println("sent") // sent pojawi się od razu
	}()

	time.Sleep(time.Second)
	<-ch
}
