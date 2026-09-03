package effectivego

import "sync"

var sem003 = make(chan int, MaxOutstanding)

func Serve2WithGate(queue chan *Request, wg *sync.WaitGroup) {
	for req := range queue {
		sem003 <- 1
		go func() {
			process(req)
			<-sem003
			wg.Done()
		}()
	}
}
