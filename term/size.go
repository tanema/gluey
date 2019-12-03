package term

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Size will return the width and height of the terminal
func Size() (width, height int, err error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return -1, -1, err
	}
	parts := strings.Split(strings.TrimRight(string(out), "\n"), " ")
	height, err = strconv.Atoi(parts[0])
	if err != nil {
		return -1, -1, err
	}
	width, err = strconv.Atoi(parts[1])
	if err != nil {
		return -1, -1, err
	}
	return width, height, nil
}
