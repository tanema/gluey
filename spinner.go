package gluey

import (
	"sync"
	"time"

	"github.com/tanema/gluey/term"
)

const spinTemplate = `
{{- range .Items -}}
	{{$.Prefix}}
	{{- if .Complete}}
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

type (
	// Spin is a single spinning status indicator
	Spin struct {
		ctx      *Ctx
		group    *SpinGroup
		Title    string
		Err      error
		Complete bool
	}
	// SpinGroup keeps a group of spinners and their statuses
	SpinGroup struct {
		ctx     *Ctx
		screen  *term.ScreenBuf
		Items   []*Spin
		current int
		on      bool
		wg      sync.WaitGroup
	}
)

// Spinner creates a single spinner and waits for it to finish
func (ctx *Ctx) Spinner(title string) *Spin {
	return ctx.NewSpinGroup().Add(title)
}

func Spinner(title string) *Spin {
	return New().Spinner(title)
}

func (spinner *Spin) Done() {
	if spinner.Complete {
		return
	}
	spinner.Complete = true
	spinner.group.render()
}

func (spinner *Spin) Fail(err error) {
	spinner.Err = err
	spinner.Done()
}

// NewSpinGroup creates a new group of spinners to track multiple statuses
func (ctx *Ctx) NewSpinGroup() *SpinGroup {
	group := &SpinGroup{ctx: ctx, screen: term.NewScreenBuf(ctx.Writer())}
	go group.run()
	return group
}

func (sg *SpinGroup) Add(title string) *Spin {
	s := &Spin{
		ctx:   sg.ctx,
		group: sg,
		Title: title,
	}
	sg.Items = append(sg.Items, s)
	return s
}

func (sg *SpinGroup) AllDone() bool {
	if len(sg.Items) == 0 {
		return false
	}
	for _, s := range sg.Items {
		if !s.Complete {
			return false
		}
	}
	return true
}

func (sg *SpinGroup) run() {
	for !sg.AllDone() {
		sg.render()
		time.Sleep(80 * time.Millisecond)
	}
	sg.render()
}

func (sg *SpinGroup) render() {
	sg.current++
	if sg.current >= len(term.SpinGlyphs) {
		sg.on = !sg.on
		sg.current = 0
	}
	data := struct {
		Glyph, Prefix string
		Items         []*Spin
		On            bool
	}{
		Glyph:  string(term.SpinGlyphs[sg.current]),
		Prefix: sg.ctx.Prefix(),
		Items:  sg.Items,
		On:     sg.on,
	}
	sg.screen.WriteTmpl(spinTemplate, data)
}
