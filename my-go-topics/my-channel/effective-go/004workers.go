package effectivego

func handle4(queue chan *Request) {
	for r := range queue {
		process(r)
	}
}

func Serve4(clientRequests chan *Request, quit chan bool) {
	// Start handlers
	for i := 0; i < MaxOutstanding; i++ {
		go handle4(clientRequests)
	}
	<-quit // Wait to be told to exit.
}
