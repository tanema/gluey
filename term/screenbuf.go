package term

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"sync"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/k0kubun/go-ansi"
)

type attribute int

type icon struct {
	color attribute
	char  string
}

const (
	fGBold      attribute = 1
	fGFaint     attribute = 2
	fGItalic    attribute = 3
	fGUnderline attribute = 4

	fGBlack   attribute = 30
	fGRed     attribute = 31
	fGGreen   attribute = 32
	fGYellow  attribute = 33
	fGBlue    attribute = 94
	fGMagenta attribute = 35
	fGCyan    attribute = 36
	fGWhite   attribute = 97
)

const (
	bGBlack attribute = iota + 40
	bGRed
	bGGreen
	bGYellow
	bGBlue
	bGMagenta
	bGCyan
	bGWhite
)

var funcMap = template.FuncMap{
	"black":   styler(fGBlack),
	"red":     styler(fGRed),
	"green":   styler(fGGreen),
	"yellow":  styler(fGYellow),
	"blue":    styler(fGBlue),
	"magenta": styler(fGMagenta),
	"cyan":    styler(fGCyan),
	"white":   styler(fGWhite),

	"bgBlack":   styler(bGBlack),
	"bgRed":     styler(bGRed),
	"bgGreen":   styler(bGGreen),
	"bgYellow":  styler(bGYellow),
	"bgBlue":    styler(bGBlue),
	"bgMagenta": styler(bGMagenta),
	"bgCyan":    styler(bGCyan),
	"bgWhite":   styler(bGWhite),

	"bold":      styler(fGBold),
	"faint":     styler(fGFaint),
	"italic":    styler(fGItalic),
	"underline": styler(fGUnderline),

	"iconQ":    iconer(iconInitial),
	"iconGood": iconer(iconGood),
	"iconWarn": iconer(iconWarn),
	"iconBad":  iconer(iconBad),
	"iconSel":  iconer(iconSelect),
	"iconChk":  iconer(iconCheckboxCheck),
	"iconBox":  iconer(iconCheckbox),
}

func styler(attr attribute) func(interface{}) string {
	return func(v interface{}) string {
		s, ok := v.(string)
		if ok && s == ">>" {
			return fmt.Sprintf("\033[%sm", strconv.Itoa(int(attr)))
		}
		return fmt.Sprintf("\033[%sm%v%s", strconv.Itoa(int(attr)), v, "\033[0m")
	}
}

func iconer(ic icon) func() string {
	return func() string { return styler(ic.color)(ic.char) }
}

// Sprintf formats a string template and outputs console ready text
func Sprintf(in string, data interface{}) string {
	return string(renderStringTemplate(in, data))
}

func renderStringTemplate(in string, data interface{}) []byte {
	tpl, err := template.New("").Funcs(funcMap).Parse(in)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

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
	linecount := bytes.Count(s.buf.Bytes(), []byte("\n"))
	s.buf.Reset()
	ClearLines(s.buf, linecount)
}

// WriteTmpl will write a text/template out to the console, using a mutex so that
// only a single writer at a time can write. This prevents the buffer from losing
// sync with the newlines
func (s *ScreenBuf) WriteTmpl(in string, data interface{}) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.reset()
	defer s.flush()
	termWidth := Width()
	tmpl := ansiwrap(renderStringTemplate(in, data), termWidth)
	if tmpl[len(tmpl)-1] != '\n' {
		tmpl = append(tmpl, '\n')
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

// ansiwrap will wrap a byte array (add linebreak) with awareness of
// ansi character widths
func ansiwrap(str []byte, width int) []byte {
	output := []byte{}
	currentChunk := []byte{}
	currentLine := []byte{}

	for _, s := range str {
		if s == '\n' {
			currentChunk = append(currentChunk, s)
			currentLine = append(currentLine, currentChunk...)
			output = append(output, currentLine...)
			currentLine = []byte{}
			currentChunk = []byte{}
			continue
		} else if s == ' ' {
			linewidth := runeCount(append(currentLine, currentChunk...))
			if linewidth > width {
				output = append(output, append(currentLine, '\n')...)
				currentLine = currentChunk
				currentChunk = []byte{}
				continue
			}
			currentLine = append(currentLine, currentChunk...)
			currentChunk = []byte{}
		}
		currentChunk = append(currentChunk, s)
	}
	currentLine = append(currentLine, currentChunk...)
	output = append(output, currentLine...)
	return output
}

// copied from ansiwrap.
// https://github.com/manifoldco/ansiwrap/blob/master/ansiwrap.go#L193
// ansiwrap worked well but I needed a version the preserved
// spacing so I just copied this method over for acurate space counting.
// There is a major problem with this though. It is not able to count
// tab spaces
func runeCount(b []byte) int {
	l := 0
	inSequence := false
	for len(b) > 0 {
		if b[0] == '\033' {
			inSequence = true
			b = b[1:]
			continue
		}
		r, rl := utf8.DecodeRune(b)
		b = b[rl:]
		if inSequence {
			if r == 'm' {
				inSequence = false
			}
			continue
		}
		if !unicode.IsPrint(r) {
			continue
		}
		l++
	}
	return l
}
