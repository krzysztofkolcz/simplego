package main

import (
	"strconv"

	effectivego001semaphore "github.com/krzysztofkolcz/my-channel/effective-go-001-semaphore"
)

func main() {

	// lesson001.Test()
	// lesson002.TestBuffered()
	// lesson002.TestUnbufferedBlocking2()
	// lesson003.TestSwitch()
	//	lesson004.TestCancel()

	// wg := &sync.WaitGroup{}
	// ch := make(chan *effectivego001semaphore.Request)
	// go effectivego001semaphore.Serve(ch, wg)
	// for i := range 10 {
	// 	wg.Add(1)
	// 	ch <- effectivego001semaphore.NewRequest(strconv.Itoa(i))
	// }
	// wg.Wait()

	// wt := make(chan int)
	// go effectivego001semaphore.Serve4(wt)
	// <-wt

	chan6 := make(chan *effectivego001semaphore.Request)
	end := make(chan int)
	go effectivego001semaphore.Serve6(chan6, end)
	for i := 0; i < 100; i++ {
		r := effectivego001semaphore.NewRequest(strconv.Itoa(i))
		chan6 <- r
	}
	end <- 1
}
