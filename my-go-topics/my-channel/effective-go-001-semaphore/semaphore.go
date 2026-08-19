package effectivego001semaphore

import (
	"fmt"
	"sync"
	"time"
)

const MaxOutstanding = 3

type Request struct {
	name string
}

func NewRequest(name string) *Request {
	return &Request{
		name: name,
	}
}

var sem = make(chan int, MaxOutstanding)

func process(r *Request) {
	fmt.Printf("request: %s \n", r.name)
	time.Sleep(1 * time.Millisecond)
}

func handle(r *Request, wg *sync.WaitGroup) {
	defer wg.Done()
	sem <- 1   // Wait for active queue to drain.
	process(r) // May take a long time.
	<-sem      // Done; enable next request to run.
}

func Serve(queue chan *Request, wg *sync.WaitGroup) {
	for {
		req := <-queue
		go handle(req, wg) // Don't wait for handle to finish.
	}
}

func Serve2WithGate(queue chan *Request, wg *sync.WaitGroup) {
	for req := range queue {
		sem <- 1
		go func() {
			process(req)
			<-sem
			wg.Done()
		}()
	}
}

func handle3(queue chan *Request) {
	for r := range queue {
		process(r)
	}
}

func Serve3(clientRequests chan *Request, quit chan bool) {
	// Start handlers
	for i := 0; i < MaxOutstanding; i++ {
		go handle3(clientRequests)
	}
	<-quit // Wait to be told to exit.
}
