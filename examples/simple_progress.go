package main

import (
	"time"

	"github.com/tanema/gluey"
)

func main() {
	bar := gluey.Progress("Loading...", 100)
	for i := 0; i <= 100; i++ {
		bar.Tick(1)
		time.Sleep(time.Millisecond)
	}
	bar.Done()
}
