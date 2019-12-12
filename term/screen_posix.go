// +build !windows

package term

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ClearLines will move the cursor up and clear the line out for re-rendering
func ClearLines(out io.Writer, linecount int) {
	out.Write([]byte(strings.Repeat("\x1b[0G\x1b[1A\x1b[0K", linecount)))
}

const (
	defaultTermWidth  = 80
	defaultTermHeight = 60
)

func size() (width, height int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return defaultTermWidth, defaultTermHeight
	}
	parts := strings.Split(strings.TrimRight(string(out), "\n"), " ")
	height, err = strconv.Atoi(parts[0])
	if err != nil {
		return defaultTermWidth, defaultTermHeight
	}
	width, err = strconv.Atoi(parts[1])
	if err != nil {
		return defaultTermWidth, defaultTermHeight
	}
	return width, height
}

// Width returns the column width of the terminal
func Width() int {
	w, _ := size()
	return w
}

// Height returns the row size of the terminal
func Height() int {
	_, h := size()
	return h
}
