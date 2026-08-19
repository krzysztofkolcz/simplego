package effectivego001semaphore

import (
	"fmt"
	"time"
)

const MaxRange = 100
const MaxSem = 10

func process4(i int) {
	time.Sleep(1 * time.Millisecond)
	fmt.Printf("process %v finished\n", i)
	<-rangesem
}

var rangesem = make(chan int, MaxSem)

func Serve4(wt chan<- int) {
	for i := range MaxRange {
		rangesem <- 1
		go process4(i)
	}
	wt <- 1
}
