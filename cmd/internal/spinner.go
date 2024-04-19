package internal

import (
	"fmt"
	"time"
)

func spinner(stopChan chan struct{}) {
	var stop bool
	delay := 100 * time.Millisecond
	go func() {
		_, ok := <-stopChan
		if ok {
			stop = true
		}
	}()
	for !stop {
		for _, r := range "-\\|/" {
			fmt.Printf("\r%c ", r)
			time.Sleep(delay)
		}
	}
}
