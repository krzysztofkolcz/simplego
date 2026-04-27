package lesson004

import (
	"context"
	"fmt"
	"time"
)

func TestCancel(){
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
	}
}