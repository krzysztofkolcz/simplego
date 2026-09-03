package effectivego

import (
	"sync"
)

var sem = make(chan int, MaxOutstanding)

func handle(r *Request, wg *sync.WaitGroup) {
	defer wg.Done()
	sem <- 1   // Wait for active queue to drain. // czyli rozumiem, że blokuje się, gdy kanał pełny
	process(r) // May take a long time.
	<-sem      // Done; enable next request to run.
}

func Serve(queue chan *Request, wg *sync.WaitGroup) {
	for {
		req := <-queue
		go handle(req, wg) // Don't wait for handle to finish.
	}
}

/* Problem z tą implementacją - Serve może tworzyć nieskończoną liczbę go rutyn, które wiszą,
czekając na swoją kolej*/
