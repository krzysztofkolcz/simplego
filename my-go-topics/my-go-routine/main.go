package main

import (
	"fmt"
	"sync"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()        // sygnalizuje zakończenie
    fmt.Printf("worker %d startuje\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go worker(i, &wg)    // uruchom goroutine
    }

    wg.Wait()              // czekaj aż wszystkie skończą
    fmt.Println("wszystkie zakończone")
}