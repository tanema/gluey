package term

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	defaultTermWidth  = 80
	defaultTermHeight = 60
)

// Size will return the width and height of the terminal
func Size() (width, height int) {
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
