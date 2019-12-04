package term

import (
	"bytes"
	"io"
	"strings"
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
	return &ScreenBuf{buf: &bytes.Buffer{}, w: w}
}

func (s *ScreenBuf) reset() {
	reset := strings.Repeat("\033[1A\033[2K\r", bytes.Count(s.buf.Bytes(), []byte("\n")))
	s.buf.Reset()
	s.buf.WriteString(reset)
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

func (s *ScreenBuf) flush() {
	io.Copy(s.w, bytes.NewBuffer(s.buf.Bytes()))
}
