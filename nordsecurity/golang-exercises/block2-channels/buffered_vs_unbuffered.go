// Package block2 zawiera ćwiczenia z Bloku 2 (channels).
package block2

import "fmt"

// Zadanie 24 — unbuffered vs buffered channel.
//
// DemoUnbuffered pokazuje kolejność zdarzeń przy unbuffered channel.
// Uruchom to (np. z małego main albo `go test -run DemoUnbuffered -v`,
// zaraz dopiszemy test) i PRZED uruchomieniem zgadnij: w jakiej kolejności
// wypiszą się te cztery linie?
func DemoUnbuffered() {
	ch := make(chan int)

	go func() {
		fmt.Println("A: goroutine zaraz wyśle do kanału")
		ch <- 1
		fmt.Println("B: goroutine — wysłanie WRÓCIŁO (ktoś już odebrał)")
	}()

	fmt.Println("C: main zaraz odbierze z kanału")
	v := <-ch
	fmt.Println("D: main odebrał", v)
}

// DemoBuffered pokazuje, że wysłanie do buforowanego kanału (z wolnym miejscem
// w buforze) NIE blokuje — nie potrzeba żadnej drugiej goroutine, żeby odebrała
// równolegle. Kolejność wypisania A, B, C jest tu w 100% deterministyczna,
// bo wszystko dzieje się w jednej goroutine, jedna linia po drugiej.
func DemoBuffered() {
	ch := make(chan int, 1)

	fmt.Println("A: main zaraz wyśle do buforowanego kanału")
	ch <- 1
	fmt.Println("B: main — wysłanie WRÓCIŁO od razu (bufor miał wolne miejsce)")
	v := <-ch
	fmt.Println("C: main odebrał", v)
}
