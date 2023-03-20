package gluey

import (
	"sync"
	"time"

	"github.com/tanema/gluey/term"
)

const spinTemplate = `
{{- range .Items -}}
	{{$.Prefix}}
	{{- if .Done}}
		{{- if .Err -}}
			{{iconBad}}
		{{- else -}}
			{{iconGood}}
		{{- end -}}
	{{- else if $.On -}}
		{{$.Glyph|cyan}}
	{{- else -}}
		{{$.Glyph}}
	{{- end}} {{.Title}}
{{end}}`

type Spinner struct {
	ctx   *Ctx
	Title string
	Err   error
	Done  bool
}

// SpinGroup keeps a group of spinners and their statuses
type SpinGroup struct {
	ctx     *Ctx
	Items   []*Spinner
	current int
	on      bool
	wg      sync.WaitGroup
}

// Spinner creates a single spinner and waits for it to finish
func (ctx *Ctx) Spinner(title string, fn func(*Spinner) error) error {
	group := ctx.NewSpinGroup()
	group.Go(title, fn)
	return group.Wait()[title]
}

// NewSpinGroup creates a new group of spinners to track multiple statuses
func (ctx *Ctx) NewSpinGroup() *SpinGroup {
	return &SpinGroup{ctx: ctx}
}

// Go adds another process to the spin group
func (sg *SpinGroup) Go(title string, fn func(*Spinner) error) {
	sg.wg.Add(1)
	s := &Spinner{ctx: sg.ctx, Title: title}
	sg.Items = append(sg.Items, s)
	go func() {
		defer sg.wg.Done()
		s.Err = fn(s)
		s.Done = true
	}()
}

// Wait will pause until all spinners are complete
func (sg *SpinGroup) Wait() map[string]error {
	done := false

	sb := term.NewScreenBuf(sg.ctx.Writer())
	defer sb.Done()

	go func() {
		for !done {
			sg.next()
			sg.render(sb)
			time.Sleep(80 * time.Millisecond)
		}
	}()
	sg.wg.Wait()
	done = true
	sg.render(sb)
	return sg.errors()
}

func (sg *SpinGroup) errors() map[string]error {
	errs := map[string]error{}
	for _, spinner := range sg.Items {
		if spinner.Err != nil {
			errs[spinner.Title] = spinner.Err
		}
	}
	return errs
}

func (sg *SpinGroup) next() {
	sg.current++
	if sg.current >= len(term.SpinGlyphs) {
		sg.on = !sg.on
		sg.current = 0
	}
}

func (sg *SpinGroup) render(sb *term.ScreenBuf) {
	data := struct {
		Glyph, Prefix string
		Items         []*Spinner
		On            bool
	}{
		Glyph:  string(term.SpinGlyphs[sg.current]),
		Prefix: sg.ctx.Prefix(),
		Items:  sg.Items,
		On:     sg.on,
	}
	sb.WriteTmpl(spinTemplate, data)
}
