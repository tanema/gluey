package promptui

import (
	"sync"
	"time"

	"github.com/tanema/promptui/term"
)

var glyphs = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

const spinTemplate = `{{- range .Items -}}
{{$.Prefix}}{{if .Done}}{{if .Err}}{{ iconBad }}{{else}}{{ iconGood }}{{end}}{{else if $.On}}{{$.CurrentGlyph | cyan}}{{else}}{{$.CurrentGlyph}}{{end}} {{ .Title }}
{{end}}`

type spinner struct {
	ctx   *Ctx
	Title string
	Err   error
	Done  bool
}

// SpinGroup keeps a group of spinners and their statuses
type SpinGroup struct {
	ctx     *Ctx
	Items   []*spinner
	current int
	On      bool
	wg      sync.WaitGroup
}

// Spinner creates a single spinner and waits for it to finish
func (ctx *Ctx) Spinner(title string, fn func() error) error {
	group := ctx.NewSpinGroup()
	group.Go(title, fn)
	return group.Wait()
}

// NewSpinGroup creates a new group of spinners to track multiple statuses
func (ctx *Ctx) NewSpinGroup() *SpinGroup {
	return &SpinGroup{ctx: ctx}
}

// Go adds another process to the spin group
func (sg *SpinGroup) Go(title string, fn func() error) {
	sg.wg.Add(1)
	s := &spinner{ctx: sg.ctx, Title: title}
	sg.Items = append(sg.Items, s)
	go func() {
		defer sg.wg.Done()
		s.Err = fn()
		s.Done = true
	}()
}

// Wait will pause until all spinners are complete
func (sg *SpinGroup) Wait() error {
	done := false
	sb := term.NewScreenBuf(sg.ctx.Writer())
	go func() {
		for !done {
			sg.next()
			sg.render(sb)
			time.Sleep(50 * time.Millisecond)
		}
	}()
	sg.wg.Wait()
	done = true
	sg.render(sb)
	return nil
}

func (sg *SpinGroup) next() {
	sg.current++
	if sg.current >= len(glyphs) {
		sg.On = !sg.On
		sg.current = 0
	}
}

func (sg *SpinGroup) CurrentGlyph() string {
	return string(glyphs[sg.current])
}

func (sg *SpinGroup) Prefix() string {
	return sg.ctx.Prefix()
}

func (sg *SpinGroup) render(sb *term.ScreenBuf) {
	sb.WriteTmpl(spinTemplate, sg)
}
