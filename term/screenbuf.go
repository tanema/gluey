package term

import (
	"bytes"
	"github.com/k0kubun/go-ansi"
	"io"
	"sync"
)

// ScreenBuf is a convenient way to write to terminal screens. It creates,
// clears and, moves up or down lines as needed to write the output to the
// terminal using ANSI escape codes.
type ScreenBuf struct {
	w   io.Writer
	buf *bytes.Buffer
	mut sync.Mutex
}

// NewScreenBuf creates and initializes a new ScreenBuf.
func NewScreenBuf(w io.Writer) *ScreenBuf {
	ansi.CursorHide()
	return &ScreenBuf{buf: &bytes.Buffer{}, w: w}
}

func (s *ScreenBuf) reset() {
	for i := 0; i < bytes.Count(s.buf.Bytes(), []byte("\n")); i++ {
		ansi.CursorPreviousLine(0)
		ansi.EraseInLine(3)
	}
	s.buf.Reset()
}

// WriteTmpl will write a text/template out to the console, using a mutex so that
// only a single writer at a time can write. This prevents the buffer from losing
// sync with the newlines
func (s *ScreenBuf) WriteTmpl(in string, data interface{}) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.reset()
	defer s.flush()
	tmpl := renderStringTemplate(in, data)
	if tmpl[len(tmpl)-1] != byte('\n') {
		tmpl = append(tmpl, []byte("\n")...)
	}
	s.buf.Write(tmpl)
}

// Done will show the cursor again and give back control
func (s *ScreenBuf) Done() {
	ansi.CursorShow()
}

func (s *ScreenBuf) flush() {
	io.Copy(s.w, bytes.NewBuffer(s.buf.Bytes()))
}
