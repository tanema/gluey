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

type barType string

const (
	barOpen   = "┏"
	barClose  = "┗"
	barDivide = "┣"
)

// Frame is a box around output that can be nested
type Frame struct {
	ctx        *Ctx
	nestedCtx  *Ctx
	timed      bool
	color      string
	closeTitle string
}

func newFrame(ctx *Ctx) *Frame {
	nestedCtx := &Ctx{Indent: ctx.Indent + 2}
	frame := &Frame{ctx: ctx, nestedCtx: nestedCtx}
	frame.SetColor("cyan")
	return frame
}

func (frame *Frame) run(title string, fn FrameFunc) error {
	frame.printBar(barOpen, title, "")
	start := time.Now()
	err := fn(frame.nestedCtx, frame)
	elapsed := time.Since(start)
	elapsedLabel := ""
	if frame.timed {
		elapsedLabel = fmt.Sprintf("(%s)", elapsed.Round(time.Second))
	}
	frame.printBar(barClose, frame.closeTitle, elapsedLabel)
	return err
}

// Divider adds a ┣━━━━ divider to the output
func (frame *Frame) Divider(label, color string) {
	frame.SetColor(color)
	frame.printBar(barDivide, label, "")
}

// SetCloseTitle sets a label that will show on the closing divider
func (frame *Frame) SetCloseTitle(label string) {
	frame.closeTitle = label
}

// SetShowElapsed enabled/disables showing the elapsed time when the frame is
// closed
func (frame *Frame) SetShowElapsed(show bool) {
	frame.timed = show
}

// SetColor will set the frames color from this point onward.
func (frame *Frame) SetColor(color string) {
	if color == "" {
		return
	}
	frame.color = color
	prefix := frame.ctx.Prefix() + Fmt("{{. | "+color+"}} ", "┃")
	frame.nestedCtx.Logger = log.New(frame.ctx.Writer(), prefix, 0)
}

func (frame *Frame) printBar(bt barType, left, right string) {
	frame.ctx.Println(frame.bar(bt, left, right))
}

func (frame *Frame) bar(bt barType, left, right string) string {
	prefix := string(bt)
	if len(left) > 0 {
		left = " " + strings.TrimSpace(left) + " "
	}
	if len(right) > 0 {
		right = " " + strings.TrimSpace(right) + " "
	}
	padding := term.Width() - len(prefix) - len(left) - len(right) - (frame.ctx.Indent - 2)
	bar := strings.Repeat("━", padding)
	return Fmt("{{ . | "+frame.color+" }}", prefix+left+bar+right)
}
