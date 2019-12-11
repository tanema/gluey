package term

import (
	"regexp"
)

const esc = "\x1b"

// PrintingWidth ANSI escape sequences (like \x1b[31m) have zero width.
// when calculating the padding width, we must exclude them.
// This also implements a basic version of utf8 character width calculation like
// we could get for real from something like utf8proc.
func PrintingWidth(str string) int {
	return len(StripCodes(str))
}

// StripCodes strips ANSI codes from a str
func StripCodes(str string) string {
	return regexp.MustCompile(`\x1b\[[\d;]+[A-z]|\r`).ReplaceAllString(str, "")
}

func control(args, cmd string) string {
	return esc + "[" + args + cmd
}

// ColorStart generates a color code https://en.wikipedia.org/wiki/ANSI_escape_code#graphics
func ColorStart(params string) string {
	return control(params, "m")
}

// ShowCursor shows the cursor
func ShowCursor() string {
	return control("", "?25h")
}

// HideCursor hide the cursor
func HideCursor() string {
	return control("", "?25l")
}

// PreviousLine move to the previous line
func PreviousLine() string {
	return control("1", "A") + control("1", "G")
}

// ClearToEndOfLine will clear from the cursor position to the end
func ClearToEndOfLine() string {
	return control("", "K")
}
