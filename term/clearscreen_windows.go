package term

import (
	"io"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
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

func clearLine(out io.Writer) {
	handle := syscall.Handle(out.Fd())

	var csbi consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi)))

	var w uint32
	csbi.cursorPosition.X = 0
	csbi.cursorPosition.Y++

	procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))))
	procFillConsoleOutputCharacter.Call(uintptr(handle), uintptr(' '), uintptr(csbi.size.X), uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))), uintptr(unsafe.Pointer(&w)))
}
