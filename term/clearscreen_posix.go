// +build !windows

package term

import (
	"io"
	"strings"
)

// ClearLines will move the cursor up and clear the line out for re-rendering
func ClearLines(out io.Writer, linecount int) {
	out.Write([]byte(strings.Repeat("\x1b[0G\x1b[1A\x1b[0K", linecount)))
}
