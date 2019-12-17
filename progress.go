package gluey

import (
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tanema/gluey/term"
)

const progressTemplate = `
{{- range .Items -}}
	{{.Prefix}}{{.Title}}{{.DoneBar|cyan}}{{.RestBar}} {{.Percent}}%
{{ end }}`

// ProgressGroup tracks a group of progress bars
type ProgressGroup struct {
	ctx   *Ctx
	Items []*Bar
	wg    sync.WaitGroup
}

// Progress creates a singel progress bar
func (ctx *Ctx) Progress(total float64, fn func(*Ctx, *Bar) error) error {
	group := &ProgressGroup{ctx: ctx}
	group.Go("", total, fn)
	return group.Wait()[""]
}

// NewProgressGroup will create a new progress bar group the will track multiple bars
func (ctx *Ctx) NewProgressGroup() *ProgressGroup {
	return &ProgressGroup{ctx: ctx}
}

// Go will add another bar to the group
func (pg *ProgressGroup) Go(title string, max float64, fn func(*Ctx, *Bar) error) {
	pg.wg.Add(1)
	if title != "" {
		title += " "
	}
	s := &Bar{ctx: pg.ctx, Title: title, total: max}
	pg.Items = append(pg.Items, s)
	go func() {
		defer pg.wg.Done()
		s.err = fn(pg.ctx, s)
		s.done = true
	}()
}

// Wait will pause until all of the progress bars are complete
func (pg *ProgressGroup) Wait() map[string]error {
	done := false
	sb := term.NewScreenBuf(pg.ctx.Writer())
	defer sb.Done()

	go func() {
		for !done {
			pg.render(sb)
			time.Sleep(50 * time.Millisecond)
		}
	}()
	pg.wg.Wait()
	done = true
	pg.render(sb)
	return pg.errors()
}

func (pg *ProgressGroup) errors() map[string]error {
	errs := map[string]error{}
	for _, bar := range pg.Items {
		if bar.err != nil {
			errs[bar.Title] = bar.err
		}
	}
	return errs
}

func (pg *ProgressGroup) render(sb *term.ScreenBuf) {
	sb.WriteTmpl(progressTemplate, pg)
}

// Bar is a single progress bar
type Bar struct {
	ctx     *Ctx
	Title   string
	DoneBar string
	RestBar string
	Prefix  string
	Percent string
	current float64
	total   float64
	err     error
	done    bool
}

// Tick allows to increment the value of the bar
func (bar *Bar) Tick(inc float64) {
	bar.set(bar.current + inc)
}

// Set allows to set the current value of the bar
func (bar *Bar) Set(val float64) {
	bar.set(val)
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
}
