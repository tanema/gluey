package term

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
)

type coord struct {
	x, y int16
}

type smallRect struct {
	left, top, right, bottom int16
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

// ClearLines will move the cursor up and clear the line out for re-rendering
func ClearLines(out io.Writer, linecount int) {
	for i := 0; i < linecount; i++ {
		clearLine(out)
	}
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

func clearLine(out io.Writer) {
	var w uint32
	csbi := termInfo()
	csbi.cursorPosition.x = 0
	csbi.cursorPosition.y--
	handle := syscall.Handle(os.Stdout.Fd())
	procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))))
	procFillConsoleOutputCharacter.Call(uintptr(handle), uintptr(' '), uintptr(csbi.size.x), uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))), uintptr(unsafe.Pointer(&w)))
}

func termInfo() consoleScreenBufferInfo {
	var csbi consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(uintptr(syscall.Handle(os.Stdout.Fd())), uintptr(unsafe.Pointer(&csbi)))
	return csbi
}

func size() (int, int) {
	csbi := termInfo()
	return int(csbi.size.x - 1), int(csbi.size.y - 1)
}
