package main

import (
	"fmt"
	"time"

	"github.com/tanema/gluey"
)

func main() {
	spinner := gluey.Spinner("testing")
	for i := 0; i <= 10; i++ {
		spinner.Title = fmt.Sprintf("testing %v/10", i)
		time.Sleep(100 * time.Millisecond)
	}
	spinner.Done()
}
