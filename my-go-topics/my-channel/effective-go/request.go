package effectivego

import (
	"fmt"
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

func process(r *Request) {
	fmt.Printf("request: %s \n", r.name)
	time.Sleep(1 * time.Millisecond)
}
