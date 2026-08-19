package effectivego001semaphore

import (
	"fmt"
	"strconv"
)

var rsem = make(chan int, 10)

func Serve5(queue chan Request) {
	for range queue {
		rsem <- 1
		go func() {
			fmt.Printf("processed\n")
			<-rsem
		}()
	}
}

const MaxHandlers6 = 6

func handler6(queue chan *Request, name string) {
	for req := range queue {
		fmt.Printf("handler: %v, process:", name)
		process(req)
	}
}

func Serve6(queue chan *Request, end chan int) {
	for i := range MaxHandlers6 {
		go handler6(queue, strconv.Itoa(i))
	}
	fmt.Printf("wait for end")
	<-end
}
