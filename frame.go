package gluey

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tanema/gluey/term"
)

// FrameFunc is the function call that is called inside the frame
type FrameFunc func(*Ctx, *Frame) error

// Frame is a box around output that can be nested
type Frame struct {
	ctx       *Ctx
	nestedCtx *Ctx
	color     string
}

func newFrame(ctx *Ctx) *Frame {
	nestedCtx := &Ctx{Indent: ctx.Indent + 2}
	frame := &Frame{ctx: ctx, nestedCtx: nestedCtx}
	frame.SetColor("cyan")
	return frame
}

func (frame *Frame) run(title string, timed bool, fn FrameFunc) error {
	frame.bar("┏", title)
	start := time.Now()
	err := fn(frame.nestedCtx, frame)
	elapsed := time.Since(start)
	closedLabel := ""
	if timed {
		closedLabel = fmt.Sprintf("%s", elapsed.Round(time.Second))
	}
	frame.bar("┗", closedLabel)
	return err
}

// Divider adds a ┣━━━━ divider to the output
func (frame *Frame) Divider(label, color string) {
	frame.SetColor(color)
	frame.bar("┣", label)
}

// SetColor will set the frames color
func (frame *Frame) SetColor(color string) {
	if color == "" {
		return
	}
	frame.color = color
	prefix := frame.ctx.Prefix() + Fmt("{{. | "+color+"}} ", "┃")
	frame.nestedCtx.Logger = log.New(frame.ctx.Writer(), prefix, 0)
}

func (frame *Frame) bar(prefix, label string) {
	if len(label) > 0 {
		label = strings.TrimSpace(label)
		label = " " + label + " "
	}
	padding := term.Width() - len(label) - len(prefix) - (frame.ctx.Indent - 2)
	bar := strings.Repeat("━", padding)
	frame.ctx.Println(Fmt("{{ . | "+frame.color+" }}", prefix+label+bar))
}
