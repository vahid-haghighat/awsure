package internal

import (
	"fmt"
	"time"
)

func spinner(stopChan chan struct{}) {
	clearLine := "\r\033[K"
	scannerFrames := `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
	var stop bool
	delay := 100 * time.Millisecond
	go func() {
		_, ok := <-stopChan
		if ok {
			stop = true
		}
	}()
	for !stop {
		for _, r := range scannerFrames {
			fmt.Printf("%s%c ", clearLine, r)
			time.Sleep(delay)
		}
	}
}
