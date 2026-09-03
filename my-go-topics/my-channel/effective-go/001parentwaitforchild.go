package effectivego

import "time"

func Parentwaitforchild() {
	c := make(chan int) // Allocate a channel.
	// Start the sort in a goroutine; when it completes, signal on the channel.
	go func() {
		// list.Sort()
		doSomething()
		c <- 1 // Send a signal; value does not matter.
	}()
	doSomethingForAWhile()
	<-c // Wait for sort to finish; discard sent value.
}

func doSomething() {
	time.Sleep(10 * time.Millisecond)
}

func doSomethingForAWhile() {
	time.Sleep(5 * time.Millisecond)
}
