package effectivego001semaphoreexercise

import "fmt"

const MaxWorkers = 10

type Request struct {
	name string
}

func NewRequest(name string) *Request {
	return &Request{
		name: name,
	}
}

func process(req Request) {
	fmt.Printf("%v", req.name)
}

func worker(req chan Request) {
	for r := range req {
		process(r)
	}

}

func Serve3(req chan Request, quit chan int) {
	for range MaxWorkers {
		go worker(req)
	}
	<-quit
}
