package term

import (
	"regexp"
	"strconv"
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

// Control returns an ANSI control sequence
func Control(args, cmd string) string {
	return esc + "[" + args + cmd
}

// Sgr generates a color code https://en.wikipedia.org/wiki/ANSI_escape_code#graphics
func Sgr(params string) string {
	return Control(params, "m")
}

// CursorUp move the cursor up n lines
func CursorUp(n int) string {
	return Control(strconv.Itoa(n), "A")
}

// CursorDown moves the cursor down n lines
func CursorDown(n int) string {
	return Control(strconv.Itoa(n), "B")
}

// CursorForward moves the cursor down n lines
func CursorForward(n int) string {
	return Control(strconv.Itoa(n), "C")
}

// CursorBack moves the cursor back n columns
func CursorBack(n int) string {
	return Control(strconv.Itoa(n), "D")
}

// CursorHorizontalAbsolute moves the cursor to a specific column
func CursorHorizontalAbsolute(n int) string {
	return Control(strconv.Itoa(n), "G")
}

// ShowCursor shows the cursor
func ShowCursor() string {
	return Control("", "?25h")
}

// HideCursor hide the cursor
func HideCursor() string {
	return Control("", "?25l")
}

// CursorSave saves the cursor position
func CursorSave() string {
	return Control("", "s")
}

// CursorRestore restores the saved cursor position
func CursorRestore() string {
	return Control("", "u")
}

// NextLine moves to the next line
func NextLine() string {
	return CursorDown(1) + CursorHorizontalAbsolute(1)
}

// PreviousLine move to the previous line
func PreviousLine() string {
	return CursorUp(1) + CursorHorizontalAbsolute(1)
}

// ClearToEndOfLine will clear from the cursor position to the end
func ClearToEndOfLine() string {
	return Control("", "K")
}
