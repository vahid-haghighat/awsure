package internal

import (
	"fmt"
	"time"
)

func spinner(stopChan chan struct{}) {
	clearLine := "\r\033[K"
	scannerFrames := []rune(`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`)
	frameCount := len(scannerFrames)
	delay := 100 * time.Millisecond
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	frameIndex := 0
	for {
		select {
		case <-stopChan:
			// Clear the line when stopping
			fmt.Printf("%s", clearLine)
			return
		case <-ticker.C:
			// Print the spinner frame
			fmt.Printf("%s%c", clearLine, scannerFrames[frameIndex])
			frameIndex = (frameIndex + 1) % frameCount
		}
	}
}
