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

func clearLine(out io.Writer) {
	handle := syscall.Handle(os.Stdout.Fd())

	var csbi consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi)))

	var w uint32
	csbi.cursorPosition.x = 0
	csbi.cursorPosition.y--

	procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))))
	procFillConsoleOutputCharacter.Call(uintptr(handle), uintptr(' '), uintptr(csbi.size.x), uintptr(*(*int32)(unsafe.Pointer(&csbi.cursorPosition))), uintptr(unsafe.Pointer(&w)))
}

func lockEcho() {
	var oldState int16
	syscall.Syscall(procGetConsoleMode.Addr(), 2, uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&oldState)), 0)
	syscall.Syscall(procSetConsoleMode.Addr(), 2, uintptr(syscall.Stdout), uintptr(oldState&(^(0x0002 | 0x0004))), 0)
}

func unlockEcho() {
	syscall.Syscall(procSetConsoleMode.Addr(), 2, uintptr(syscall.Stdout), uintptr(oldState), 0)
}
