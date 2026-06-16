package gluey

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/tanema/gluey/term"
)

const progressTemplate = `
{{- range .Items -}}
	{{.Prefix}}{{.Title}}{{.DoneBar|cyan}}{{.RestBar}} {{.Percent}}%
{{ end }}`

type (
	// Bar is a single progress bar
	Bar struct {
		ctx     *Ctx
		group   *ProgressGroup
		Title   string
		DoneBar string
		RestBar string
		Prefix  string
		Percent string
		current float64
		total   float64
		err     error
		done    bool
		mut     sync.Mutex
	}
	// ProgressGroup tracks a group of progress bars
	ProgressGroup struct {
		ctx    *Ctx
		screen *term.ScreenBuf
		Items  []*Bar
		wg     sync.WaitGroup
	}
)

func Progress(title string, total float64) *Bar {
	return New().Progress(title, total)
}

// Progress creates a singel progress bar
func (ctx *Ctx) Progress(title string, total float64) *Bar {
	return ctx.NewProgressGroup().Add(title, total)
}

// NewProgressGroup will create a new progress bar group the will track multiple bars
func (ctx *Ctx) NewProgressGroup() *ProgressGroup {
	return &ProgressGroup{ctx: ctx, screen: term.NewScreenBuf(ctx.Writer())}
}

// Add will add another bar to the group
func (pg *ProgressGroup) Add(title string, max float64) *Bar {
	pg.wg.Add(1)
	if title != "" {
		title += " "
	}
	s := &Bar{ctx: pg.ctx, group: pg, Title: title, total: max}
	pg.Items = append(pg.Items, s)
	pg.render()
	return s
}

func (pg *ProgressGroup) AllDone() bool {
	for _, s := range pg.Items {
		if !s.done {
			return false
		}
	}
	return true
}

func (pg *ProgressGroup) render() {
	if pg.screen == nil {
		return
	}
	pg.screen.WriteTmpl(progressTemplate, pg)
}

// Tick allows to increment the value of the bar
func (bar *Bar) Tick(inc float64) {
	bar.mut.Lock()
	defer bar.mut.Unlock()
	bar.set(bar.current + inc)
}

// Set allows to set the current value of the bar
func (bar *Bar) Set(val float64) {
	bar.mut.Lock()
	defer bar.mut.Unlock()
	bar.set(val)
}

func (bar *Bar) Done() {
	bar.mut.Lock()
	defer bar.mut.Unlock()
	bar.done = true
	bar.set(bar.total)
}

func (bar *Bar) Fail(err error) {
	bar.err = err
	bar.Done()
}

func (bar *Bar) set(val float64) {
	bar.current = math.Max(0, math.Min(val, bar.total))
	bar.done = bar.current == bar.total
	bar.Percent = strconv.Itoa(int((bar.current / bar.total) * 100))

	percent := bar.current / bar.total
	barwidth := term.Width() - (bar.ctx.Indent - 2) - len(bar.Title) - len(bar.Percent) - 4
	done := percent * float64(barwidth)
	bar.DoneBar = strings.Repeat("█", int(done))
	bar.RestBar = strings.Repeat("░", int(math.Max(float64(barwidth)-done, 0)))
	bar.Prefix = bar.ctx.Prefix()
	bar.group.render()
}
